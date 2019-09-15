package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources"
	"github.com/toggler-io/toggler/services/release"
	"github.com/toggler-io/toggler/services/security"
	_ "github.com/lib/pq"
)

func NewPostgres(db *sql.DB) (*Postgres, error) {
	pg := &Postgres{DB: db}

	if err := Migrate(db); err != nil {
		return nil, err
	}

	return pg, nil
}

type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Postgres struct {
	DB
	timeLocation *time.Location
}

func (pg *Postgres) Close() error {
	switch db := pg.DB.(type) {
	case *sql.DB:
		return db.Close()
	case *sql.Tx:
		return db.Commit()
	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) Save(ctx context.Context, entity interface{}) error {
	if currentID, ok := resources.LookupID(entity); !ok || currentID != "" {
		return fmt.Errorf("entity already have an ID: %s", currentID)
	}

	switch e := entity.(type) {
	case *release.Flag:
		return pg.featureFlagInsertNew(ctx, e)
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
	case *release.Flag:
		return pg.featureFlagFindByID(ctx, e, id)

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

func (pg *Postgres) Truncate(ctx context.Context, Type interface{}) error {
	var tableName string
	switch Type.(type) {
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
	case release.Flag, *release.Flag:
		query = `DELETE FROM "release_flags" WHERE id = $1`

	case release.Pilot, *release.Pilot:
		query = `DELETE FROM "pilots" WHERE id = $1`

	case security.Token, *security.Token:
		query = `DELETE FROM "tokens" WHERE id = $1`

	case resources.TestEntity, *resources.TestEntity:
		query = `DELETE FROM "test_entities" WHERE id = $1`

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
	case *release.Flag:
		return pg.featureFlagUpdate(ctx, e)

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
	case release.Flag, *release.Flag:
		return pg.featureFlagFindAll(ctx)

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

	mapper := featureFlagMapper{}
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

const tokenFindByTokenStringQuery = `
SELECT id, sha512, duration, issued_at, owner_uid
FROM "tokens" 
WHERE sha512 = $1;
`

func (pg *Postgres) FindTokenBySHA512Hex(ctx context.Context, token string) (*security.Token, error) {
	row := pg.DB.QueryRowContext(ctx, tokenFindByTokenStringQuery, token)
	var t security.Token

	err := row.Scan(
		&t.ID,
		&t.SHA512,
		&t.Duration,
		&t.IssuedAt,
		&t.OwnerUID,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if t.IssuedAt, err = pg.ensureTimeLocation(t.IssuedAt); err != nil {
		return &t, err
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

func (pg *Postgres) FindFlagsByName(ctx context.Context, flagNames ...string) frameless.Iterator {

	var namesInClause []string
	var args []interface{}

	namesInClause = append(namesInClause)
	for i, arg := range flagNames {
		namesInClause = append(namesInClause, fmt.Sprintf(`$%d`, i+1))
		args = append(args, arg)
	}

	mapper := featureFlagMapper{}

	query := fmt.Sprintf(`%s FROM "release_flags" WHERE "name" IN (%s)`,
		mapper.SelectClause(),
		strings.Join(namesInClause, `,`))

	flags, err := pg.DB.QueryContext(ctx, query, args...)

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(flags, mapper)
}

const featureFlagInsertNewQuery = `
INSERT INTO "release_flags" (
	name,
	rollout_rand_seed,
	rollout_strategy_percentage,
	rollout_strategy_decision_logic_api
)
VALUES ($1, $2, $3, $4)
RETURNING id;
`

func (pg *Postgres) featureFlagInsertNew(ctx context.Context, flag *release.Flag) error {

	var DecisionLogicAPI sql.NullString

	if flag.Rollout.Strategy.DecisionLogicAPI != nil {
		DecisionLogicAPI.Valid = true
		DecisionLogicAPI.String = flag.Rollout.Strategy.DecisionLogicAPI.String()
	}

	row := pg.DB.QueryRowContext(ctx, featureFlagInsertNewQuery,
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

func (pg *Postgres) featureFlagFindByID(ctx context.Context, flag *release.Flag, id int64) (bool, error) {

	mapper := featureFlagMapper{}
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

const tokenFindByIDQuery = `
SELECT id, sha512, duration, issued_at, owner_uid
FROM "tokens" 
WHERE id = $1;
`

func (pg *Postgres) tokenFindByID(ctx context.Context, token *security.Token, id int64) (bool, error) {
	row := pg.DB.QueryRowContext(ctx, tokenFindByIDQuery, id)
	var t security.Token

	err := row.Scan(
		&t.ID,
		&t.SHA512,
		&t.Duration,
		&t.IssuedAt,
		&t.OwnerUID,
	)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if pg.timeLocation == nil {
		if pg.timeLocation, err = time.LoadLocation(`UTC`); err != nil {
			return false, err
		}
	}

	if t.IssuedAt, err = pg.ensureTimeLocation(t.IssuedAt); err != nil {
		return false, err
	}

	*token = t
	return true, nil
}

func (pg *Postgres) featureFlagFindAll(ctx context.Context) frameless.Iterator {
	mapper := featureFlagMapper{}
	query := fmt.Sprintf(`%s FROM "release_flags"`, mapper.SelectClause())
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
SELECT id, sha512, duration, issued_at, owner_uid
FROM "tokens"
`

func (pg *Postgres) tokenFindAll(ctx context.Context) frameless.Iterator {
	rows, err := pg.DB.QueryContext(ctx, tokenFindAllQuery)

	if err != nil {
		return iterators.NewError(err)
	}

	receiver, sender := iterators.NewPipe()

	go func() {
		defer sender.Close()

	wrk:
		for rows.Next() {

			var t security.Token

			err := rows.Scan(
				&t.ID,
				&t.SHA512,
				&t.Duration,
				&t.IssuedAt,
				&t.OwnerUID,
			)

			if err == sql.ErrNoRows {
				break wrk
			}

			if err != nil {
				sender.Error(err)
				break wrk
			}

			if t.IssuedAt, err = pg.ensureTimeLocation(t.IssuedAt); err != nil {
				sender.Error(err)
				break wrk
			}

			if err := sender.Encode(&t); err != nil {
				sender.Error(err)
				break wrk
			}
		}

		if err := rows.Err(); err != nil {
			sender.Error(err)
		}

	}()

	return receiver
}

const featureFlagUpdateQuery = `
UPDATE "release_flags"
SET name = $1,
    rollout_rand_seed = $2,
    rollout_strategy_percentage = $3,
    rollout_strategy_decision_logic_api = $4
WHERE id = $5;
`

func (pg *Postgres) featureFlagUpdate(ctx context.Context, flag *release.Flag) error {
	var DecisionLogicAPI sql.NullString

	if flag.Rollout.Strategy.DecisionLogicAPI != nil {
		DecisionLogicAPI.Valid = true
		DecisionLogicAPI.String = flag.Rollout.Strategy.DecisionLogicAPI.String()
	}

	_, err := pg.DB.ExecContext(ctx, featureFlagUpdateQuery,
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

func (pg *Postgres) ensureTimeLocation(t time.Time) (time.Time, error) {
	if pg.timeLocation == nil {
		var err error

		if pg.timeLocation, err = time.LoadLocation(`UTC`); err != nil {
			return t, err
		}
	}

	return t.In(pg.timeLocation), nil
}

type featureFlagMapper struct{}

func (featureFlagMapper) SelectClause() string {
	return `SELECT id, name, rollout_rand_seed, rollout_strategy_percentage, rollout_strategy_decision_logic_api`
}

func (featureFlagMapper) Map(scanner iterators.SQLRowScanner, ptr frameless.Entity) error {
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

	return reflects.Link(&ff, ptr)
}

type pilotMapper struct{}

func (pilotMapper) SelectClause() string {
	const query = `id, feature_flag_id, external_id, enrolled`
	return query
}

func (pilotMapper) Map(s iterators.SQLRowScanner, ptr frameless.Entity) error {
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

	return reflects.Link(&p, ptr)
}
