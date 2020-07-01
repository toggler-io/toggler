package storages

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/errs"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	_ "github.com/lib/pq"

	pgmigr "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/toggler-io/toggler/domains/deployment"
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
	DB *sql.DB
}

type PostgresTxCtxKey struct{}

func (pg *Postgres) BeginTx(ctx context.Context) (context.Context, error) {
	if ctx.Value(PostgresTxCtxKey{}) != nil {
		return nil, fmt.Errorf(`context has already a postgres tx`)
	}

	tx, err := pg.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, PostgresTxCtxKey{}, tx), nil
}

func (pg *Postgres) CommitTx(ctx context.Context) error {
	tx, ok := ctx.Value(PostgresTxCtxKey{}).(*sql.Tx)
	if !ok {
		return fmt.Errorf(`no postgres tx in the given context`)
	}

	return tx.Commit()
}

func (pg *Postgres) RollbackTx(ctx context.Context) error {
	tx, ok := ctx.Value(PostgresTxCtxKey{}).(*sql.Tx)
	if !ok {
		return fmt.Errorf(`no postgres tx in the given context`)
	}

	return tx.Rollback()
}

func (pg *Postgres) getDB(ctx context.Context) interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
} {
	if tx, ok := ctx.Value(PostgresTxCtxKey{}).(*sql.Tx); ok {
		return tx
	}

	return pg.DB
}

func (pg *Postgres) FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx context.Context, flag release.Flag, env deployment.Environment, rollout *release.Rollout) (bool, error) {
	var m releaseRolloutMapper
	tmpl := `SELECT %s FROM release_rollouts WHERE flag_id = $1 AND env_id = $2`
	query := fmt.Sprintf(tmpl, strings.Join(m.Columns(), `, `))
	row := pg.getDB(ctx).QueryRowContext(ctx, query, flag.ID, env.ID)

	err := m.Map(row, rollout)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (pg *Postgres) FindDeploymentEnvironmentByAlias(ctx context.Context, idOrName string, env *deployment.Environment) (bool, error) {
	var (
		format string
		query  string
		m      deploymentEnvironmentMapper
	)
	if isUUIDValid(idOrName) {
		format = `SELECT %s FROM deployment_environments WHERE id = $1`
	} else {
		format = `SELECT %s FROM deployment_environments WHERE name = $1`
	}
	query = fmt.Sprintf(format, strings.Join(m.Columns(), `,`))
	err := m.Map(pg.getDB(ctx).QueryRowContext(ctx, query, idOrName), env)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (pg *Postgres) Close() error {
	return pg.DB.Close()
}

func (pg *Postgres) Create(ctx context.Context, ptr interface{}) error {
	if currentID, ok := resources.LookupID(ptr); !ok || currentID != "" {
		return fmt.Errorf("entity already have an ID: %s", currentID)
	}

	switch e := ptr.(type) {
	case *deployment.Environment:
		return pg.deploymentEnvironmentInsertNew(ctx, e)
	case *release.Flag:
		return pg.releaseFlagInsertNew(ctx, e)
	case *release.Rollout:
		return pg.releaseRolloutInsertNew(ctx, e)
	case *release.ManualPilot:
		return pg.pilotInsertNew(ctx, e)
	case *security.Token:
		return pg.tokenInsertNew(ctx, e)
	case *resources.TestEntity:
		return pg.testEntityInsertNew(ctx, e)
	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) FindByID(ctx context.Context, ptr interface{}, id string) (bool, error) {
	if !isUUIDValid(id) {
		return false, nil
	}

	switch e := ptr.(type) {
	case *deployment.Environment:
		return pg.deploymentEnvironmentFindByID(ctx, e, id)
	case *release.Flag:
		return pg.releaseFlagFindByID(ctx, e, id)
	case *release.Rollout:
		return pg.releaseRolloutFindByID(ctx, e, id)
	case *release.ManualPilot:
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
	case deployment.Environment, *deployment.Environment:
		tableName = `deployment_environments`
	case release.Flag, *release.Flag:
		tableName = `release_flags`
	case release.Rollout, *release.Rollout:
		tableName = `release_rollouts`
	case release.ManualPilot, *release.ManualPilot:
		tableName = `release_pilots`
	case security.Token, *security.Token:
		tableName = `tokens`
	case resources.TestEntity, *resources.TestEntity:
		tableName = `test_entities`
	default:
		return frameless.ErrNotImplemented
	}

	query := fmt.Sprintf(`DELETE FROM "%s"`, tableName)
	_, err := pg.getDB(ctx).ExecContext(ctx, query)
	return err
}

func (pg *Postgres) DeleteByID(ctx context.Context, Type interface{}, id string) error {
	if !isUUIDValid(id) {
		return frameless.ErrNotFound
	}

	var query string
	switch Type.(type) {
	case deployment.Environment, *deployment.Environment:
		query = `DELETE FROM "deployment_environments" WHERE "id" = $1`

	case release.Flag, *release.Flag:
		query = `DELETE FROM "release_flags" WHERE "id" = $1`

	case release.Rollout, *release.Rollout:
		query = `DELETE FROM "release_rollouts" WHERE "id" = $1`

	case release.ManualPilot, *release.ManualPilot:
		query = `DELETE FROM "release_pilots" WHERE "id" = $1`

	case security.Token, *security.Token:
		query = `DELETE FROM "tokens" WHERE "id" = $1`

	case resources.TestEntity, *resources.TestEntity:
		query = `DELETE FROM "test_entities" WHERE "id" = $1`

	default:
		return frameless.ErrNotImplemented
	}

	result, err := pg.getDB(ctx).ExecContext(ctx, query, id)
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
	case *deployment.Environment:
		return pg.deploymentEnvironmentUpdate(ctx, e)

	case *release.Flag:
		return pg.releaseFlagUpdate(ctx, e)

	case *release.Rollout:
		return pg.releaseRolloutUpdate(ctx, e)

	case *release.ManualPilot:
		return pg.pilotUpdate(ctx, e)

	case *security.Token:
		return pg.tokenUpdate(ctx, e)

	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) FindAll(ctx context.Context, Type interface{}) frameless.Iterator {
	switch Type.(type) {
	case deployment.Environment, *deployment.Environment:
		return pg.deploymentEnvironmentFindAll(ctx)
	case release.Flag, *release.Flag:
		return pg.releaseFlagFindAll(ctx)
	case release.Rollout, *release.Rollout:
		return pg.releaseRolloutFindAll(ctx)
	case release.ManualPilot, *release.ManualPilot:
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

	row := pg.getDB(ctx).QueryRowContext(ctx, query, name)

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

func (pg *Postgres) FindReleaseManualPilotByExternalID(ctx context.Context, flagID, envID, pilotExtID string) (*release.ManualPilot, error) {
	if !isUUIDValid(flagID) {
		return nil, nil
	}

	m := pilotMapper{}
	q := fmt.Sprintf(`SELECT %s FROM "release_pilots" WHERE "flag_id" = $1 AND "env_id" = $2 AND "external_id" = $3`,
		strings.Join(m.Columns(), `, `))

	row := pg.getDB(ctx).QueryRowContext(ctx, q, flagID, envID, pilotExtID)

	var p release.ManualPilot

	err := m.Map(row, &p)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (pg *Postgres) FindReleasePilotsByReleaseFlag(ctx context.Context, flag release.Flag) release.PilotEntries {
	if flag.ID == `` {
		return iterators.NewEmpty()
	}

	if flag.ID == `` {
		return iterators.NewEmpty()
	}

	if !isUUIDValid(flag.ID) {
		return iterators.NewEmpty()
	}

	m := pilotMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "release_pilots" WHERE "flag_id" = $1`, strings.Join(m.Columns(), `, `))
	rows, err := pg.getDB(ctx).QueryContext(ctx, query, flag.ID)

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

	row := pg.getDB(ctx).QueryRowContext(ctx, tokenFindByTokenStringQuery, token)

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

func (pg *Postgres) FindReleasePilotsByExternalID(ctx context.Context, pilotExtID string) release.PilotEntries {
	m := pilotMapper{}
	q := fmt.Sprintf(`SELECT %s FROM "release_pilots" WHERE "external_id" = $1`, strings.Join(m.Columns(), `, `))
	rows, err := pg.getDB(ctx).QueryContext(ctx, q, pilotExtID)
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

	flags, err := pg.getDB(ctx).QueryContext(ctx, query, args...)

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(flags, mapper)
}

const releaseFlagInsertNewQuery = `
INSERT INTO "release_flags" (id, name)
VALUES ($1, $2);
`

func (pg *Postgres) releaseFlagInsertNew(ctx context.Context, flag *release.Flag) error {
	id, err := newV4UUID()
	if err != nil {
		return err
	}

	if _, err := pg.getDB(ctx).ExecContext(ctx, releaseFlagInsertNewQuery, id, flag.Name,
	); err != nil {
		return err
	}

	return resources.SetID(flag, id)
}

const releaseRolloutInsertNewQuery = `
INSERT INTO "release_rollouts" (id, flag_id, env_id, plan)
VALUES ($1, $2, $3, $4);
`

func (pg *Postgres) releaseRolloutInsertNew(ctx context.Context, rollout *release.Rollout) error {
	id, err := newV4UUID()
	if err != nil {
		return err
	}

	planJSON, err := json.Marshal(release.RolloutDefinitionView{Definition: rollout.Plan})

	if err != nil {
		return err
	}

	if _, err := pg.getDB(ctx).ExecContext(ctx, releaseRolloutInsertNewQuery,
		id,
		rollout.FlagID,
		rollout.DeploymentEnvironmentID,
		planJSON,
	); err != nil {
		return err
	}

	return resources.SetID(rollout, id)
}

const deploymentEnvironmentInsertNewQuery = `
INSERT INTO "deployment_environments" (id, name)
VALUES ($1, $2);
`

func (pg *Postgres) deploymentEnvironmentInsertNew(ctx context.Context, env *deployment.Environment) error {
	id, err := newV4UUID()
	if err != nil {
		return err
	}

	if _, err := pg.getDB(ctx).ExecContext(ctx, deploymentEnvironmentInsertNewQuery, id, env.Name); err != nil {
		return err
	}

	return resources.SetID(env, id)
}

const pilotInsertNewQuery = `
INSERT INTO "release_pilots" (id, flag_id, env_id, external_id, is_participating)
VALUES ($1, $2, $3, $4, $5);
`

func (pg *Postgres) pilotInsertNew(ctx context.Context, pilot *release.ManualPilot) error {

	id, err := newV4UUID()
	if err != nil {
		return err
	}

	if !isUUIDValid(pilot.FlagID) {
		return fmt.Errorf(`invalid name Flag ID: ` + pilot.FlagID)
	}

	if _, err := pg.getDB(ctx).ExecContext(ctx, pilotInsertNewQuery,
		id,
		pilot.FlagID,
		pilot.DeploymentEnvironmentID,
		pilot.ExternalID,
		pilot.IsParticipating,
	); err != nil {
		return err
	}

	return resources.SetID(pilot, id)

}

const testEntityInsertNewQuery = `
INSERT INTO "test_entities" (id) 
VALUES ($1)
RETURNING id;
`

func (pg *Postgres) testEntityInsertNew(ctx context.Context, te *resources.TestEntity) error {
	id, err := newV4UUID()
	if err != nil {
		return err
	}

	if _, err := pg.getDB(ctx).ExecContext(ctx, testEntityInsertNewQuery, id); err != nil {
		return err
	}

	return resources.SetID(te, id)
}

const tokenInsertNewQuery = `
INSERT INTO "tokens" (id, sha512, owner_uid, issued_at, duration)
VALUES ($1, $2, $3, $4, $5);
`

func (pg *Postgres) tokenInsertNew(ctx context.Context, token *security.Token) error {

	id, err := newV4UUID()
	if err != nil {
		return err
	}

	if _, err := pg.getDB(ctx).ExecContext(ctx, tokenInsertNewQuery,
		id,
		token.SHA512,
		token.OwnerUID,
		token.IssuedAt,
		token.Duration,
	); err != nil {
		return err
	}

	return resources.SetID(token, id)

}

func (pg *Postgres) releaseRolloutFindByID(ctx context.Context, rollout *release.Rollout, id string) (bool, error) {
	mapper := releaseRolloutMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "release_rollouts" WHERE "id" = $1`, strings.Join(mapper.Columns(), `, `))
	err := mapper.Map(pg.getDB(ctx).QueryRowContext(ctx, query, id), rollout)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (pg *Postgres) releaseFlagFindByID(ctx context.Context, flag *release.Flag, id string) (bool, error) {
	mapper := releaseFlagMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "release_flags" WHERE "id" = $1`, strings.Join(mapper.Columns(), `, `))
	err := mapper.Map(pg.getDB(ctx).QueryRowContext(ctx, query, id), flag)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (pg *Postgres) deploymentEnvironmentFindByID(ctx context.Context, env *deployment.Environment, id string) (bool, error) {
	mapper := deploymentEnvironmentMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "deployment_environments" WHERE "id" = $1`, strings.Join(mapper.Columns(), `, `))
	err := mapper.Map(pg.getDB(ctx).QueryRowContext(ctx, query, id), env)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (pg *Postgres) pilotFindByID(ctx context.Context, pilot *release.ManualPilot, id string) (bool, error) {
	m := pilotMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "release_pilots" WHERE "id" = $1`, strings.Join(m.Columns(), `, `))
	row := pg.getDB(ctx).QueryRowContext(ctx, query, id)

	var p release.ManualPilot
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

func (pg *Postgres) testEntityFindByID(ctx context.Context, testEntity *resources.TestEntity, id string) (bool, error) {
	row := pg.getDB(ctx).QueryRowContext(ctx, testEntityFindByIDQuery, id)

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

func (pg *Postgres) tokenFindByID(ctx context.Context, token *security.Token, id string) (bool, error) {

	row := pg.getDB(ctx).QueryRowContext(ctx, tokenFindByIDQuery, id)
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
	rows, err := pg.getDB(ctx).QueryContext(ctx, query)
	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, mapper)
}

func (pg *Postgres) releaseRolloutFindAll(ctx context.Context) frameless.Iterator {
	mapper := releaseRolloutMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "release_rollouts"`, strings.Join(mapper.Columns(), `, `))
	rows, err := pg.getDB(ctx).QueryContext(ctx, query)
	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, mapper)
}

func (pg *Postgres) deploymentEnvironmentFindAll(ctx context.Context) frameless.Iterator {
	mapper := deploymentEnvironmentMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "deployment_environments"`, strings.Join(mapper.Columns(), `,`))

	rows, err := pg.getDB(ctx).QueryContext(ctx, query)
	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, mapper)
}

func (pg *Postgres) pilotFindAll(ctx context.Context) frameless.Iterator {
	m := pilotMapper{}
	q := fmt.Sprintf(`SELECT %s FROM "release_pilots"`, strings.Join(m.Columns(), `, `))
	rows, err := pg.getDB(ctx).QueryContext(ctx, q)
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

	rows, err := pg.getDB(ctx).QueryContext(ctx, `SELECT id FROM "test_entities"`)

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

	rows, err := pg.getDB(ctx).QueryContext(ctx, fmt.Sprintf(tokenFindAllQuery, strings.Join(m.Columns(), `, `)))

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, m)
}

const deploymentEnvironmentUpdateQuery = `
UPDATE "deployment_environments"
SET name = $2
WHERE id = $1;
`

func (pg *Postgres) deploymentEnvironmentUpdate(ctx context.Context, env *deployment.Environment) error {
	_, err := pg.getDB(ctx).ExecContext(ctx, deploymentEnvironmentUpdateQuery, env.ID, env.Name)

	return err
}

const releaseFlagUpdateQuery = `
UPDATE "release_flags"
SET name = $2
WHERE id = $1;
`

func (pg *Postgres) releaseFlagUpdate(ctx context.Context, flag *release.Flag) error {
	_, err := pg.getDB(ctx).ExecContext(ctx, releaseFlagUpdateQuery, flag.ID, flag.Name)

	return err
}

const releaseRolloutUpdateQuery = `
UPDATE "release_rollouts"
SET flag_id = $2, 
    env_id = $3, 
    plan = $4
WHERE id = $1;
`

func (pg *Postgres) releaseRolloutUpdate(ctx context.Context, rollout *release.Rollout) error {
	planJSON, err := json.Marshal(release.RolloutDefinitionView{Definition: rollout.Plan})
	if err != nil {
		return err
	}

	_, err = pg.getDB(ctx).ExecContext(ctx, releaseRolloutUpdateQuery,
		rollout.ID,
		rollout.FlagID,
		rollout.DeploymentEnvironmentID,
		planJSON,
	)

	return err
}

const pilotUpdateQuery = `
UPDATE "release_pilots"
SET flag_id = $2,
	env_id = $3,
    external_id = $4,
    is_participating = $5
WHERE id = $1;
`

func (pg *Postgres) pilotUpdate(ctx context.Context, pilot *release.ManualPilot) error {
	_, err := pg.getDB(ctx).ExecContext(ctx, pilotUpdateQuery, pilot.ID,
		pilot.FlagID,
		pilot.DeploymentEnvironmentID,
		pilot.ExternalID,
		pilot.IsParticipating,
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
	_, err := pg.getDB(ctx).ExecContext(ctx, tokenUpdateQuery,
		t.SHA512,
		t.OwnerUID,
		t.IssuedAt,
		t.Duration,
		t.ID,
	)

	return err
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

type deploymentEnvironmentMapper struct{}

func (m deploymentEnvironmentMapper) Columns() []string {
	return []string{`id`, `name`}
}

func (m deploymentEnvironmentMapper) Map(s iterators.SQLRowScanner, ptr interface{}) error {
	var src deployment.Environment
	if err := s.Scan(
		&src.ID,
		&src.Name,
	); err != nil {
		return err
	}
	return reflects.Link(src, ptr)
}

type releaseRolloutMapper struct{}

func (releaseRolloutMapper) Columns() []string {
	return []string{`id`, `flag_id`, `env_id`, `plan`}
}

type releaseRolloutPlanValue struct {
	release.RolloutDefinition
}

func (rp releaseRolloutPlanValue) Value() (driver.Value, error) {
	return json.Marshal(release.RolloutDefinitionView{Definition: rp.RolloutDefinition})
}

func (rp *releaseRolloutPlanValue) Scan(iSRC interface{}) error {
	src, ok := iSRC.([]byte)
	if !ok {
		const err errs.Error = "Type assertion .([]byte) failed."
		return err
	}

	var rpv release.RolloutDefinitionView
	if err := json.Unmarshal(src, &rpv); err != nil {
		return err
	}

	rp.RolloutDefinition = rpv.Definition
	return nil
}

func (releaseRolloutMapper) Map(scanner iterators.SQLRowScanner, ptr interface{}) error {
	var rollout release.Rollout

	var rolloutPlanValue releaseRolloutPlanValue

	if err := scanner.Scan(
		&rollout.ID,
		&rollout.FlagID,
		&rollout.DeploymentEnvironmentID,
		&rolloutPlanValue,
	); err != nil {
		return err
	}

	rollout.Plan = rolloutPlanValue.RolloutDefinition
	return reflects.Link(rollout, ptr)
}

type releaseFlagMapper struct{}

func (releaseFlagMapper) SelectClause() string {
	return `SELECT id, name`
}

func (releaseFlagMapper) Columns() []string {
	return []string{`id`, `name`}
}

func (releaseFlagMapper) Map(scanner iterators.SQLRowScanner, ptr interface{}) error {
	var flag release.Flag
	if err := scanner.Scan(&flag.ID, &flag.Name); err != nil {
		return err
	}
	return reflects.Link(flag, ptr)
}

type pilotMapper struct{}

func (pilotMapper) Columns() []string {
	return []string{
		`id`,
		`flag_id`,
		`env_id`,
		`external_id`,
		`is_participating`,
	}
}

func (pilotMapper) Map(s iterators.SQLRowScanner, ptr interface{}) error {
	var p release.ManualPilot

	err := s.Scan(
		&p.ID,
		&p.FlagID,
		&p.DeploymentEnvironmentID,
		&p.ExternalID,
		&p.IsParticipating,
	)

	if err != nil {
		return err
	}

	return reflects.Link(p, ptr)
}
