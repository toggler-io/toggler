package storages

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	_ "github.com/lib/pq"

	pgmigr "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/external/resource/storages/migrations"
)

func NewPostgres(db *sql.DB) (*Postgres, error) {
	pg := &Postgres{DB: db}

	if err := PostgresMigrate(db); err != nil {
		return nil, err
	}

	return pg, nil
}

type Postgres struct {
	DB interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	}
}

func (pg *Postgres) FindReleaseAllowsByReleaseFlags(ctx context.Context, flags ...release.Flag) release.AllowEntries {
	m := releaseAllowMapper{}
	nextPlaceholder := newPrepareQueryPlaceholderAssigner()

	var query string
	var args []interface{}

	var queryWhereFlagInList []string
	for _, f := range flags {
		queryWhereFlagInList = append(queryWhereFlagInList, nextPlaceholder())
		args = append(args, f.ID)
	}

	query = fmt.Sprintf(`SELECT %s FROM "release_flag_ip_addr_allows" WHERE "flag_id" IN (%s)`,
		strings.Join(m.Columns(), `, `),
		strings.Join(queryWhereFlagInList, `, `))

	rows, err := pg.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, m)
}

func (pg *Postgres) Close() error {
	switch db := pg.DB.(type) {
	case interface{ Close() error }:
		return db.Close()
	case interface{ Commit() error }:
		return db.Commit()
	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) Create(ctx context.Context, ptr interface{}) error {
	if currentID, ok := resources.LookupID(ptr); !ok || currentID != "" {
		return fmt.Errorf("entity already have an ID: %s", currentID)
	}

	switch e := ptr.(type) {
	case *release.IPAllow:
		return pg.releaseAllowInsertNew(ctx, e)
	case *release.Flag:
		return pg.releaseFlagInsertNew(ctx, e)
	case *release.Pilot:
		return pg.pilotInsertNew(ctx, e)
	case *security.Token:
		return pg.tokenInsertNew(ctx, e)
	case *resources.TestEntity:
		return pg.testEntityInsertNew(ctx, e)
	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) FindByID(ctx context.Context, ptr interface{}, ID string) (bool, error) {
	id, err := strconv.ParseInt(ID, 10, 64)

	if err != nil {
		return false, nil
	}

	switch e := ptr.(type) {
	case *release.IPAllow:
		return pg.releaseAllowFindByID(ctx, e, id)

	case *release.Flag:
		return pg.releaseFlagFindByID(ctx, e, id)

	case *release.Pilot:
		return pg.pilotFindByID(ctx, e, id)

	case *security.Token:
		return pg.tokenFindByID(ctx, e, id)

	case *resources.TestEntity:
		return pg.testEntityFindByID(ctx, e, id)

	default:
		return false, frameless.ErrNotImplemented
	}
}

func (pg *Postgres) DeleteAll(ctx context.Context, Type interface{}) error {
	var tableName string
	switch Type.(type) {
	case release.IPAllow, *release.IPAllow:
		tableName = `release_flag_ip_addr_allows`
	case release.Flag, *release.Flag:
		tableName = `release_flags`
	case release.Pilot, *release.Pilot:
		tableName = `pilots`
	case security.Token, *security.Token:
		tableName = `tokens`
	case resources.TestEntity, *resources.TestEntity:
		tableName = `test_entities`
	default:
		return frameless.ErrNotImplemented
	}

	query := fmt.Sprintf(`DELETE FROM "%s"`, tableName)
	_, err := pg.DB.ExecContext(ctx, query)
	return err
}

func (pg *Postgres) DeleteByID(ctx context.Context, Type interface{}, ID string) error {
	id, err := strconv.ParseInt(ID, 10, 64)

	if err != nil {
		return frameless.ErrNotFound
	}

	var query string

	switch Type.(type) {
	case release.IPAllow, *release.IPAllow:
		query = `DELETE FROM "release_flag_ip_addr_allows" WHERE "id" = $1`

	case release.Flag, *release.Flag:
		query = `DELETE FROM "release_flags" WHERE "id" = $1`

	case release.Pilot, *release.Pilot:
		query = `DELETE FROM "pilots" WHERE "id" = $1`

	case security.Token, *security.Token:
		query = `DELETE FROM "tokens" WHERE "id" = $1`

	case resources.TestEntity, *resources.TestEntity:
		query = `DELETE FROM "test_entities" WHERE "id" = $1`

	default:
		return frameless.ErrNotImplemented
	}

	result, err := pg.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return frameless.ErrNotFound
	}

	return nil
}

func (pg *Postgres) Update(ctx context.Context, ptr interface{}) error {
	switch e := ptr.(type) {
	case *release.IPAllow:
		return pg.releaseAllowUpdate(ctx, e)

	case *release.Flag:
		return pg.releaseFlagUpdate(ctx, e)

	case *release.Pilot:
		return pg.pilotUpdate(ctx, e)

	case *security.Token:
		return pg.tokenUpdate(ctx, e)

	case *resources.TestEntity:
		return pg.testEntityUpdate(ctx, e)

	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) FindAll(ctx context.Context, Type interface{}) frameless.Iterator {
	switch Type.(type) {
	case release.IPAllow, *release.IPAllow:
		return pg.releaseAllowFindAll(ctx)

	case release.Flag, *release.Flag:
		return pg.releaseFlagFindAll(ctx)

	case release.Pilot, *release.Pilot:
		return pg.pilotFindAll(ctx)

	case security.Token, *security.Token:
		return pg.tokenFindAll(ctx)

	case resources.TestEntity, *resources.TestEntity:
		return pg.testEntityFindAll(ctx)

	default:
		return iterators.NewError(frameless.ErrNotImplemented)
	}
}

func (pg *Postgres) FindReleaseFlagByName(ctx context.Context, name string) (*release.Flag, error) {

	mapper := releaseFlagMapper{}
	query := fmt.Sprintf(`%s FROM "release_flags" WHERE "name" = $1`,
		mapper.SelectClause())

	row := pg.DB.QueryRowContext(ctx, query, name)

	var ff release.Flag

	err := mapper.Map(row, &ff)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &ff, nil

}

func (pg *Postgres) FindReleaseFlagPilotByPilotExternalID(ctx context.Context, FeatureFlagID, ExternalPilotID string) (*release.Pilot, error) {
	flagID, err := strconv.ParseInt(FeatureFlagID, 10, 64)

	if err != nil {
		return nil, nil
	}

	m := pilotMapper{}
	q := fmt.Sprintf(`SELECT %s FROM "pilots" WHERE "feature_flag_id" = $1 AND "external_id" = $2`,
		m.SelectClause())

	row := pg.DB.QueryRowContext(ctx, q, flagID, ExternalPilotID)

	var p release.Pilot

	err = m.Map(row, &p)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (pg *Postgres) FindPilotsByFeatureFlag(ctx context.Context, ff *release.Flag) frameless.Iterator {
	if ff == nil {
		return iterators.NewEmpty()
	}

	if ff.ID == `` {
		return iterators.NewEmpty()
	}

	ffID, err := strconv.ParseInt(ff.ID, 10, 64)

	if err != nil {
		return iterators.NewEmpty()
	}

	m := pilotMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "pilots" WHERE "feature_flag_id" = $1`, m.SelectClause())
	rows, err := pg.DB.QueryContext(ctx, query, ffID)

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, m)
}

const tokenFindByTokenStringTemplate = `
SELECT %s
FROM "tokens" 
WHERE sha512 = $1;
`

var tokenFindByTokenStringQuery = fmt.Sprintf(tokenFindByTokenStringTemplate, strings.Join(tokenMapper{}.Columns(), `,`))

func (pg *Postgres) FindTokenBySHA512Hex(ctx context.Context, token string) (*security.Token, error) {
	m := tokenMapper{}

	row := pg.DB.QueryRowContext(ctx, tokenFindByTokenStringQuery, token)

	var t security.Token

	err := m.Map(row, &t)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (pg *Postgres) FindPilotEntriesByExtID(ctx context.Context, pilotExtID string) release.PilotEntries {
	m := pilotMapper{}
	q := fmt.Sprintf(`SELECT %s FROM "pilots" WHERE "external_id" = $1`, m.SelectClause())
	rows, err := pg.DB.QueryContext(ctx, q, pilotExtID)
	if err != nil {
		return iterators.NewError(err)
	}
	return iterators.NewSQLRows(rows, m)
}

func (pg *Postgres) FindReleaseFlagsByName(ctx context.Context, flagNames ...string) frameless.Iterator {

	var namesInClause []string
	var args []interface{}

	namesInClause = append(namesInClause)
	for i, arg := range flagNames {
		namesInClause = append(namesInClause, fmt.Sprintf(`$%d`, i+1))
		args = append(args, arg)
	}

	mapper := releaseFlagMapper{}

	query := fmt.Sprintf(`%s FROM "release_flags" WHERE "name" IN (%s)`,
		mapper.SelectClause(),
		strings.Join(namesInClause, `,`))

	flags, err := pg.DB.QueryContext(ctx, query, args...)

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(flags, mapper)
}

const releaseAllowInsertNewQuery = `
INSERT INTO "release_flag_ip_addr_allows" ("flag_id", "ip_addr")
VALUES ($1, $2)
RETURNING "id";
`

func (pg *Postgres) releaseAllowInsertNew(ctx context.Context, allow *release.IPAllow) error {

	var ipAddr sql.NullString

	if allow.InternetProtocolAddress != `` {
		ipAddr.Valid = true
		ipAddr.String = allow.InternetProtocolAddress
	}

	if allow.FlagID == `` {
		return errors.New(`creating release.IPAllow without FlagID is forbidden`)
	}

	row := pg.DB.QueryRowContext(ctx, releaseAllowInsertNewQuery,
		allow.FlagID,
		ipAddr,
	)

	var id sql.NullInt64
	if err := row.Scan(&id); err != nil {
		return err
	}

	return resources.SetID(allow, strconv.FormatInt(id.Int64, 10))
}

const releaseFlagInsertNewQuery = `
INSERT INTO "release_flags" (
	name,
	rollout_rand_seed,
	rollout_strategy_percentage,
	rollout_strategy_decision_logic_api
)
VALUES ($1, $2, $3, $4)
RETURNING id;
`

func (pg *Postgres) releaseFlagInsertNew(ctx context.Context, flag *release.Flag) error {

	var DecisionLogicAPI sql.NullString

	if flag.Rollout.Strategy.DecisionLogicAPI != nil {
		DecisionLogicAPI.Valid = true
		DecisionLogicAPI.String = flag.Rollout.Strategy.DecisionLogicAPI.String()
	}

	row := pg.DB.QueryRowContext(ctx, releaseFlagInsertNewQuery,
		flag.Name,
		flag.Rollout.RandSeed,
		flag.Rollout.Strategy.Percentage,
		DecisionLogicAPI,
	)

	var id sql.NullInt64
	if err := row.Scan(&id); err != nil {
		return err
	}

	return resources.SetID(flag, strconv.FormatInt(id.Int64, 10))
}

const pilotInsertNewQuery = `
INSERT INTO "pilots" (feature_flag_id, external_id, enrolled)
VALUES ($1, $2, $3)
RETURNING id;
`

func (pg *Postgres) pilotInsertNew(ctx context.Context, pilot *release.Pilot) error {

	flagID, err := strconv.ParseInt(pilot.FlagID, 10, 64)

	if err != nil {
		return fmt.Errorf(`invalid Feature Flag ID: ` + pilot.FlagID)
	}

	row := pg.DB.QueryRowContext(ctx, pilotInsertNewQuery,
		flagID,
		pilot.ExternalID,
		pilot.Enrolled,
	)

	var id sql.NullInt64
	if err := row.Scan(&id); err != nil {
		return err
	}

	return resources.SetID(pilot, strconv.FormatInt(id.Int64, 10))
}

const testEntityInsertNewQuery = `
INSERT INTO "test_entities" (id) 
VALUES (default)
RETURNING id;
`

func (pg *Postgres) testEntityInsertNew(ctx context.Context, te *resources.TestEntity) error {
	row := pg.DB.QueryRowContext(ctx, testEntityInsertNewQuery)

	var id sql.NullInt64
	if err := row.Scan(&id); err != nil {
		return err
	}

	return resources.SetID(te, strconv.FormatInt(id.Int64, 10))
}

const tokenInsertNewQuery = `
INSERT INTO "tokens" (sha512, owner_uid, issued_at, duration)
VALUES ($1, $2, $3, $4)
RETURNING id;
`

func (pg *Postgres) tokenInsertNew(ctx context.Context, token *security.Token) error {
	row := pg.DB.QueryRowContext(ctx, tokenInsertNewQuery,
		token.SHA512,
		token.OwnerUID,
		token.IssuedAt,
		token.Duration,
	)

	var id sql.NullInt64
	if err := row.Scan(&id); err != nil {
		return err
	}

	return resources.SetID(token, strconv.FormatInt(id.Int64, 10))
}

func (pg *Postgres) releaseAllowFindByID(ctx context.Context, allow *release.IPAllow, id int64) (bool, error) {
	mapper := releaseAllowMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "release_flag_ip_addr_allows" WHERE "id" = $1`, strings.Join(mapper.Columns(), `, `))
	row := pg.DB.QueryRowContext(ctx, query, id)
	err := mapper.Map(row, allow)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (pg *Postgres) releaseFlagFindByID(ctx context.Context, flag *release.Flag, id int64) (bool, error) {

	mapper := releaseFlagMapper{}
	query := fmt.Sprintf(`%s FROM "release_flags" WHERE "id" = $1`, mapper.SelectClause())
	row := pg.DB.QueryRowContext(ctx, query, id)

	var ff release.Flag
	err := mapper.Map(row, &ff)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	*flag = ff
	return true, nil

}

func (pg *Postgres) pilotFindByID(ctx context.Context, pilot *release.Pilot, id int64) (bool, error) {
	m := pilotMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "pilots" WHERE "id" = $1`, m.SelectClause())
	row := pg.DB.QueryRowContext(ctx, query, id)

	var p release.Pilot
	err := m.Map(row, &p)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	*pilot = p
	return true, nil
}

const testEntityFindByIDQuery = `
SELECT id
FROM "test_entities" 
WHERE id = $1;
`

func (pg *Postgres) testEntityFindByID(ctx context.Context, testEntity *resources.TestEntity, id int64) (bool, error) {
	row := pg.DB.QueryRowContext(ctx, testEntityFindByIDQuery, id)

	err := row.Scan(&testEntity.ID)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

const tokenFindByIDQueryTemplate = `
SELECT %s
FROM "tokens" 
WHERE id = $1;
`

var tokenFindByIDQuery = fmt.Sprintf(tokenFindByIDQueryTemplate, strings.Join(tokenMapper{}.Columns(), `, `))

func (pg *Postgres) tokenFindByID(ctx context.Context, token *security.Token, id int64) (bool, error) {

	row := pg.DB.QueryRowContext(ctx, tokenFindByIDQuery, id)
	err := tokenMapper{}.Map(row, token)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (pg *Postgres) releaseFlagFindAll(ctx context.Context) frameless.Iterator {
	mapper := releaseFlagMapper{}
	query := fmt.Sprintf(`%s FROM "release_flags"`, mapper.SelectClause())
	rows, err := pg.DB.QueryContext(ctx, query)
	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, mapper)
}

func (pg *Postgres) releaseAllowFindAll(ctx context.Context) frameless.Iterator {
	mapper := releaseAllowMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "release_flag_ip_addr_allows"`, strings.Join(mapper.Columns(), `, `))
	rows, err := pg.DB.QueryContext(ctx, query)
	if err != nil {
		return iterators.NewError(err)
	}
	return iterators.NewSQLRows(rows, mapper)
}

func (pg *Postgres) pilotFindAll(ctx context.Context) frameless.Iterator {
	m := pilotMapper{}
	q := fmt.Sprintf(`SELECT %s FROM "pilots"`, m.SelectClause())
	rows, err := pg.DB.QueryContext(ctx, q)
	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, m)
}

func (pg *Postgres) testEntityFindAll(ctx context.Context) frameless.Iterator {

	mapper := iterators.SQLRowMapperFunc(func(s iterators.SQLRowScanner, e frameless.Entity) error {
		te := e.(*resources.TestEntity)
		return s.Scan(&te.ID)
	})

	rows, err := pg.DB.QueryContext(ctx, `SELECT id FROM "test_entities"`)

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, mapper)

}

const tokenFindAllQuery = `
SELECT %s
FROM "tokens"
`

func (pg *Postgres) tokenFindAll(ctx context.Context) frameless.Iterator {
	m := tokenMapper{}

	rows, err := pg.DB.QueryContext(ctx, fmt.Sprintf(tokenFindAllQuery, strings.Join(m.Columns(), `, `)))

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, m)
}

const releaseAllowUpdateQuery = `
UPDATE "release_flag_ip_addr_allows"
SET flag_id = $2,
	ip_addr = $3
WHERE id = $1;
`

func (pg *Postgres) releaseAllowUpdate(ctx context.Context, allow *release.IPAllow) error {
	_, err := pg.DB.ExecContext(ctx, releaseAllowUpdateQuery,
		allow.ID,
		allow.FlagID,
		allow.InternetProtocolAddress,
	)
	return err
}

const releaseFlagUpdateQuery = `
UPDATE "release_flags"
SET name = $1,
    rollout_rand_seed = $2,
    rollout_strategy_percentage = $3,
    rollout_strategy_decision_logic_api = $4
WHERE id = $5;
`

func (pg *Postgres) releaseFlagUpdate(ctx context.Context, flag *release.Flag) error {
	var DecisionLogicAPI sql.NullString

	if flag.Rollout.Strategy.DecisionLogicAPI != nil {
		DecisionLogicAPI.Valid = true
		DecisionLogicAPI.String = flag.Rollout.Strategy.DecisionLogicAPI.String()
	}

	_, err := pg.DB.ExecContext(ctx, releaseFlagUpdateQuery,
		flag.Name,
		flag.Rollout.RandSeed,
		flag.Rollout.Strategy.Percentage,
		DecisionLogicAPI,
		flag.ID,
	)

	return err
}

const pilotUpdateQuery = `
UPDATE "pilots"
SET feature_flag_id = $1,
    external_id = $2,
    enrolled = $3
WHERE id = $4;
`

func (pg *Postgres) pilotUpdate(ctx context.Context, pilot *release.Pilot) error {
	_, err := pg.DB.ExecContext(ctx, pilotUpdateQuery,
		pilot.FlagID,
		pilot.ExternalID,
		pilot.Enrolled,
		pilot.ID,
	)

	return err
}

const tokenUpdateQuery = `
UPDATE "tokens"
SET sha512 = $1,
    owner_uid = $2,
    issued_at = $3,
    duration = $4
WHERE id = $5;
`

func (pg *Postgres) tokenUpdate(ctx context.Context, t *security.Token) error {
	_, err := pg.DB.ExecContext(ctx, tokenUpdateQuery,
		t.SHA512,
		t.OwnerUID,
		t.IssuedAt,
		t.Duration,
		t.ID,
	)

	return err
}

func (pg *Postgres) testEntityUpdate(ctx context.Context, t *resources.TestEntity) error {
	return nil
}

func newPrepareQueryPlaceholderAssigner() func() string {
	var index int
	return func() string {
		index++
		return fmt.Sprintf(`$%d`, index)
	}
}

/* -------------------------- MIGRATION -------------------------- */

//go:generate esc -o ./migrations/fs.go -pkg migrations ./migrations
const pgMigrationsDirectory = `/migrations/postgres`

func PostgresMigrate(db *sql.DB) error {

	m, err := NewPostgresMigrate(db)

	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil

}

func NewPostgresMigrate(db *sql.DB) (*migrate.Migrate, error) {

	srcDriver, err := NewPostgresBindataSourceDriver()

	if err != nil {
		return nil, err
	}

	dbDriver, err := pgmigr.WithInstance(db, &pgmigr.Config{})

	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance(`esc`, srcDriver, `postgres`, dbDriver)

	if err != nil {
		return nil, err
	}

	return m, err

}

func NewPostgresBindataSourceDriver() (source.Driver, error) {
	f, err := migrations.FS(false).Open(pgMigrationsDirectory)

	if err != nil {
		return nil, err
	}

	fis, err := f.Readdir(-1)

	if err != nil {
		return nil, err
	}

	var names []string

	for _, fi := range fis {
		if !fi.IsDir() {
			names = append(names, fi.Name())
		}
	}

	s := bindata.Resource(names, func(name string) ([]byte, error) {
		// path.Join used for `/` file separator joining.
		// This is needed because the assets generated in an environment where "/" is used as file separator.
		return migrations.FSByte(false, path.Join(pgMigrationsDirectory, name))
	})

	return bindata.WithInstance(s)
}

type tokenMapper struct{}

func (m tokenMapper) Columns() []string {
	return []string{`id`, `sha512`, `duration`, `issued_at`, `owner_uid`}
}

func (m tokenMapper) Map(s iterators.SQLRowScanner, ptr interface{}) error {
	var src security.Token
	if err := s.Scan(
		&src.ID,
		&src.SHA512,
		&src.Duration,
		&src.IssuedAt,
		&src.OwnerUID,
	); err != nil {
		return err
	}
	src.IssuedAt = src.IssuedAt.In(timeLocation)
	return reflects.Link(src, ptr)
}

type releaseAllowMapper struct{}

func (releaseAllowMapper) Columns() []string {
	return []string{`id`, `flag_id`, `ip_addr`}
}

func (releaseAllowMapper) Map(scanner iterators.SQLRowScanner, ptr interface{}) error {
	var ipAllow release.IPAllow
	var ipAddr sql.NullString
	if err := scanner.Scan(&ipAllow.ID, &ipAllow.FlagID, &ipAddr); err != nil {
		return err
	}
	if ipAddr.Valid {
		ipAllow.InternetProtocolAddress = ipAddr.String
	}
	return reflects.Link(ipAllow, ptr)
}

type releaseFlagMapper struct{}

func (releaseFlagMapper) SelectClause() string {
	return `SELECT id, name, rollout_rand_seed, rollout_strategy_percentage, rollout_strategy_decision_logic_api`
}

func (releaseFlagMapper) Map(scanner iterators.SQLRowScanner, ptr interface{}) error {
	var ff release.Flag
	var DecisionLogicAPI sql.NullString

	err := scanner.Scan(
		&ff.ID,
		&ff.Name,
		&ff.Rollout.RandSeed,
		&ff.Rollout.Strategy.Percentage,
		&DecisionLogicAPI,
	)

	if err != nil {
		return err
	}

	if DecisionLogicAPI.Valid {
		u, err := url.ParseRequestURI(DecisionLogicAPI.String)
		if err != nil {
			return err
		}
		ff.Rollout.Strategy.DecisionLogicAPI = u
	}

	return reflects.Link(ff, ptr)
}

type pilotMapper struct{}

func (pilotMapper) SelectClause() string {
	const query = `id, feature_flag_id, external_id, enrolled`
	return query
}

func (pilotMapper) Map(s iterators.SQLRowScanner, ptr interface{}) error {
	var p release.Pilot

	err := s.Scan(
		&p.ID,
		&p.FlagID,
		&p.ExternalID,
		&p.Enrolled,
	)

	if err != nil {
		return err
	}

	return reflects.Link(p, ptr)
}
