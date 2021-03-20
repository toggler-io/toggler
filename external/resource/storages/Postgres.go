package storages

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/adamluzsi/frameless/consterror"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/lib/pq"
	_ "github.com/lib/pq"

	pgmigr "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/external/resource/storages/migrations"
)

func NewPostgres(db *sql.DB, dsn string) (*Postgres, error) {
	pg := &Postgres{DB: db, DSN: dsn}

	if err := PostgresMigrate(db); err != nil {
		return nil, err
	}

	return pg, nil
}

type Postgres struct {
	DB  *sql.DB
	DSN string
}

type PostgresTxCtxKey struct{}

type PostgresTx struct {
	depth int
	*sql.Tx
}

func (pg *Postgres) BeginTx(ctx context.Context) (context.Context, error) {
	if tx, ok := pg.lookupTx(ctx); ok && tx.Tx != nil {
		tx.depth++
		return ctx, nil
	}

	tx, err := pg.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, PostgresTxCtxKey{}, &PostgresTx{Tx: tx}), nil
}

func (pg *Postgres) CommitTx(ctx context.Context) error {
	tx, ok := pg.lookupTx(ctx)
	if !ok {
		return fmt.Errorf(`no postgres tx in the given context`)
	}

	if tx.depth > 0 {
		tx.depth--
		return nil
	}

	return tx.Commit()
}

func (pg *Postgres) RollbackTx(ctx context.Context) error {
	tx, ok := pg.lookupTx(ctx)
	if !ok {
		return fmt.Errorf(`no postgres tx in the given context`)
	}

	return tx.Rollback()
}

func (pg *Postgres) lookupTx(ctx context.Context) (*PostgresTx, bool) {
	tx, ok := ctx.Value(PostgresTxCtxKey{}).(*PostgresTx)
	return tx, ok
}

// return func meant to ensure tx is only closed if we made it
func (pg *Postgres) getTx(ctx context.Context) (tx *sql.Tx, commit func() error, rollback func() error, err error) {
	db := pg.getDB(ctx)

	switch db := db.(type) {
	case *sql.DB:
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return nil, nil, nil, err
		}
		return tx, tx.Commit, tx.Rollback, nil

	case *sql.Tx:
		return db, func() error { return nil }, func() error { return nil }, nil

	default:
		return nil, nil, nil, fmt.Errorf(`unknown db type received: %T`, db)
	}
}

func (pg *Postgres) getDB(ctx context.Context) interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
} {
	if tx, ok := pg.lookupTx(ctx); ok && tx.Tx != nil {
		return tx.Tx
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
	err := EnsureID(ptr)
	if err != nil {
		return err
	}

	if currentID, _ := resources.LookupID(ptr); !isUUIDValid(currentID) {
		return fmt.Errorf(`invalid ID: %#v`, currentID)
	}

	switch e := ptr.(type) {
	case *deployment.Environment:
		return pg.deploymentEnvironmentCreate(ctx, e)
	case *release.Flag:
		return pg.releaseFlagCreate(ctx, e)
	case *release.Rollout:
		return pg.releaseRolloutCreate(ctx, e)
	case *release.ManualPilot:
		return pg.releaseManualPilotCreate(ctx, e)
	case *security.Token:
		return pg.securityTokenCreate(ctx, e)
	default:
		return fmt.Errorf(`ErrNotImplemented`)
	}
}

func (pg *Postgres) FindByID(ctx context.Context, ptr, id interface{}) (bool, error) {
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
	default:
		return false, fmt.Errorf(`ErrNotImplemented`)
	}
}

func (pg *Postgres) DeleteAll(ctx context.Context, Type interface{}) error {
	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}

	var (
		tableName string
		topicName string
		message   interface{}
	)
	switch Type.(type) {
	case deployment.Environment, *deployment.Environment:
		tableName = `deployment_environments`
		topicName = deploymentEnvironmentDeleteAllSubscriptionName
		message = deployment.Environment{}
	case release.Flag, *release.Flag:
		tableName = `release_flags`
		topicName = releaseFlagDeleteAllSubscriptionName
		message = release.Flag{}
	case release.Rollout, *release.Rollout:
		tableName = `release_rollouts`
		topicName = releaseRolloutDeleteAllSubscriptionName
		message = release.Rollout{}
	case release.ManualPilot, *release.ManualPilot:
		tableName = `release_pilots`
		topicName = releaseManualPilotDeleteAllSubscriptionName
		message = release.ManualPilot{}
	case security.Token, *security.Token:
		tableName = `tokens`
		topicName = securityTokenDeleteAllSubscriptionName
		message = security.Token{}
	case resources.TestEntity, *resources.TestEntity:
		tableName = `test_entities`
	default:
		return fmt.Errorf(`ErrNotImplemented`)
	}

	query := fmt.Sprintf(`DELETE FROM "%s"`, tableName)
	if _, err := tx.ExecContext(ctx, query); err != nil {
		_ = rollback()
		return err
	}

	if message != nil {
		if err := pg.notify(ctx, tx, topicName, message); err != nil {
			_ = rollback()
			return err
		}
	}

	return commit()
}

func (pg *Postgres) DeleteByID(ctx context.Context, T, id interface{}) error {
	if !isUUIDValid(id) {
		return fmt.Errorf(`ErrNotFound`)
	}

	sid := id.(string)

	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}

	var (
		query     string
		topicName string
		message   interface{}
	)
	switch T.(type) {
	case deployment.Environment, *deployment.Environment:
		query = `DELETE FROM "deployment_environments" WHERE "id" = $1`
		topicName = deploymentEnvironmentDeleteByIDSubscriptionName
		message = deployment.Environment{ID: sid}

	case release.Flag, *release.Flag:
		query = `DELETE FROM "release_flags" WHERE "id" = $1`
		topicName = releaseFlagDeleteByIDSubscriptionName
		message = release.Flag{ID: sid}

	case release.Rollout, *release.Rollout:
		query = `DELETE FROM "release_rollouts" WHERE "id" = $1`
		topicName = releaseRolloutDeleteByIDSubscriptionName
		message = release.Rollout{ID: sid}

	case release.ManualPilot, *release.ManualPilot:
		query = `DELETE FROM "release_pilots" WHERE "id" = $1`
		topicName = releaseManualPilotDeleteByIDSubscriptionName
		message = release.ManualPilot{ID: sid}

	case security.Token, *security.Token:
		query = `DELETE FROM "tokens" WHERE "id" = $1`
		topicName = securityTokenDeleteByIDSubscriptionName
		message = security.Token{ID: sid}

	default:
		return fmt.Errorf(`ErrNotImplemented`)
	}

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		_ = rollback()
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		_ = rollback()
		return err
	}

	if count == 0 {
		_ = rollback()
		return fmt.Errorf(`ErrNotFound`)
	}

	if message != nil {
		if err := pg.notify(ctx, tx, topicName, message); err != nil {
			_ = rollback()
			return err
		}
	}

	return commit()
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
		return pg.releaseManualPilotUpdate(ctx, e)

	case *security.Token:
		return pg.securityTokenUpdate(ctx, e)

	default:
		return fmt.Errorf(`ErrNotImplemented`)
	}
}

func (pg *Postgres) FindAll(ctx context.Context, Type interface{}) iterators.Interface {
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
	default:
		return iterators.NewError(fmt.Errorf(`ErrNotImplemented`))
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

func (pg *Postgres) FindReleaseManualPilotByExternalID(ctx context.Context, flagID, envID interface{}, pilotExtID string) (*release.ManualPilot, error) {
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

func (pg *Postgres) FindReleaseFlagsByName(ctx context.Context, flagNames ...string) iterators.Interface {
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

func (pg *Postgres) releaseFlagCreate(ctx context.Context, flag *release.Flag) error {
	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, releaseFlagInsertNewQuery, flag.ID, flag.Name); err != nil {
		return err
	}

	if err := pg.notify(ctx, tx, releaseFlagCreateSubscriptionName, *flag); err != nil {
		_ = rollback()
		return err
	}

	return commit()
}

const releaseRolloutInsertNewQuery = `
INSERT INTO "release_rollouts" (id, flag_id, env_id, plan)
VALUES ($1, $2, $3, $4);
`

func (pg *Postgres) releaseRolloutCreate(ctx context.Context, rollout *release.Rollout) error {
	planJSON, err := json.Marshal(release.RolloutDefinitionView{Definition: rollout.Plan})

	if err != nil {
		return err
	}

	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, releaseRolloutInsertNewQuery,
		rollout.ID,
		rollout.FlagID,
		rollout.DeploymentEnvironmentID,
		planJSON,
	); err != nil {
		_ = rollback()
		return err
	}

	if err := pg.notify(ctx, tx, releaseRolloutCreateSubscriptionName, *rollout); err != nil {
		_ = rollback()
		return err
	}

	return commit()
}

const deploymentEnvironmentInsertNewQuery = `
INSERT INTO "deployment_environments" (id, name)
VALUES ($1, $2);
`

func (pg *Postgres) deploymentEnvironmentCreate(ctx context.Context, env *deployment.Environment) error {
	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, deploymentEnvironmentInsertNewQuery, env.ID, env.Name); err != nil {
		_ = rollback()
		return err
	}

	if err := pg.notify(ctx, tx, deploymentEnvironmentCreateSubscriptionName, *env); err != nil {
		_ = rollback()
		return err
	}

	return commit()
}

const pilotInsertNewQuery = `
INSERT INTO "release_pilots" (id, flag_id, env_id, external_id, is_participating)
VALUES ($1, $2, $3, $4, $5);
`

func (pg *Postgres) releaseManualPilotCreate(ctx context.Context, pilot *release.ManualPilot) error {
	if !isUUIDValid(pilot.FlagID) {
		return fmt.Errorf(`invalid name Flag ID: ` + pilot.FlagID)
	}

	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, pilotInsertNewQuery,
		pilot.ID,
		pilot.FlagID,
		pilot.DeploymentEnvironmentID,
		pilot.ExternalID,
		pilot.IsParticipating,
	); err != nil {
		_ = rollback()
		return err
	}

	if err := pg.notify(ctx, tx, releaseManualPilotCreateSubscriptionName, *pilot); err != nil {
		_ = rollback()
		return err
	}

	return commit()
}

const tokenInsertNewQuery = `
INSERT INTO "tokens" (id, sha512, owner_uid, issued_at, duration)
VALUES ($1, $2, $3, $4, $5);
`

func (pg *Postgres) securityTokenCreate(ctx context.Context, token *security.Token) error {
	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, tokenInsertNewQuery,
		token.ID,
		token.SHA512,
		token.OwnerUID,
		token.IssuedAt,
		token.Duration,
	); err != nil {
		_ = rollback()
		return err
	}

	if err := pg.notify(ctx, tx, securityTokenCreateSubscriptionName, *token); err != nil {
		_ = rollback()
		return err
	}

	return commit()
}

func (pg *Postgres) releaseRolloutFindByID(ctx context.Context, rollout *release.Rollout, id interface{}) (bool, error) {
	mapper := releaseRolloutMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "release_rollouts" WHERE "id" = $1`, strings.Join(mapper.Columns(), `, `))
	err := mapper.Map(pg.getDB(ctx).QueryRowContext(ctx, query, id), rollout)

	if err == sql.ErrNoRows {
		query = fmt.Sprintf(`SELECT %s FROM "release_rollouts"`, strings.Join(mapper.Columns(), `, `))
		rows, err := pg.getDB(ctx).QueryContext(ctx, query)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		for rows.Next() {
			var ro release.Rollout
			if err := mapper.Map(rows, &ro); err != nil {
				panic(err)
			}
		}

		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (pg *Postgres) releaseFlagFindByID(ctx context.Context, flag *release.Flag, id interface{}) (bool, error) {
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

func (pg *Postgres) deploymentEnvironmentFindByID(ctx context.Context, env *deployment.Environment, id interface{}) (bool, error) {
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

func (pg *Postgres) pilotFindByID(ctx context.Context, pilot *release.ManualPilot, id interface{}) (bool, error) {
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

const tokenFindByIDQueryTemplate = `
SELECT %s
FROM "tokens" 
WHERE id = $1;
`

var tokenFindByIDQuery = fmt.Sprintf(tokenFindByIDQueryTemplate, strings.Join(tokenMapper{}.Columns(), `, `))

func (pg *Postgres) tokenFindByID(ctx context.Context, token *security.Token, id interface{}) (bool, error) {

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

func (pg *Postgres) releaseFlagFindAll(ctx context.Context) iterators.Interface {
	mapper := releaseFlagMapper{}
	query := fmt.Sprintf(`%s FROM "release_flags"`, mapper.SelectClause())
	rows, err := pg.getDB(ctx).QueryContext(ctx, query)
	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, mapper)
}

func (pg *Postgres) releaseRolloutFindAll(ctx context.Context) iterators.Interface {
	mapper := releaseRolloutMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "release_rollouts"`, strings.Join(mapper.Columns(), `, `))
	rows, err := pg.getDB(ctx).QueryContext(ctx, query)
	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, mapper)
}

func (pg *Postgres) deploymentEnvironmentFindAll(ctx context.Context) iterators.Interface {
	mapper := deploymentEnvironmentMapper{}
	query := fmt.Sprintf(`SELECT %s FROM "deployment_environments"`, strings.Join(mapper.Columns(), `,`))

	rows, err := pg.getDB(ctx).QueryContext(ctx, query)
	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, mapper)
}

func (pg *Postgres) pilotFindAll(ctx context.Context) iterators.Interface {
	m := pilotMapper{}
	q := fmt.Sprintf(`SELECT %s FROM "release_pilots"`, strings.Join(m.Columns(), `, `))
	rows, err := pg.getDB(ctx).QueryContext(ctx, q)
	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, m)
}

const tokenFindAllQuery = `
SELECT %s
FROM "tokens"
`

func (pg *Postgres) tokenFindAll(ctx context.Context) iterators.Interface {
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

func (pg *Postgres) deploymentEnvironmentUpdate(ctx context.Context, env *deployment.Environment) (rErr error) {
	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rErr != nil {
			_ = rollback()
			return
		}
		rErr = commit()
	}()

	if res, err := tx.ExecContext(ctx, deploymentEnvironmentUpdateQuery, env.ID, env.Name); err != nil {
		return err
	} else {
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return fmt.Errorf(`deployment environment not found`)
		}
	}

	return pg.notify(ctx, tx, deploymentEnvironmentUpdateSubscriptionName, *env)
}

const releaseFlagUpdateQuery = `
UPDATE "release_flags"
SET name = $2
WHERE id = $1;
`

func (pg *Postgres) releaseFlagUpdate(ctx context.Context, flag *release.Flag) (rErr error) {
	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rErr != nil {
			_ = rollback()
			return
		}
		rErr = commit()
	}()

	if res, err := tx.ExecContext(ctx, releaseFlagUpdateQuery, flag.ID, flag.Name); err != nil {
		return err
	} else {
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return fmt.Errorf(`release flag not found`)
		}
	}

	return pg.notify(ctx, tx, releaseFlagUpdateSubscriptionName, *flag)
}

const releaseRolloutUpdateQuery = `
UPDATE "release_rollouts"
SET flag_id = $2, 
    env_id = $3, 
    plan = $4
WHERE id = $1;
`

func (pg *Postgres) releaseRolloutUpdate(ctx context.Context, rollout *release.Rollout) (rErr error) {
	planJSON, err := json.Marshal(release.RolloutDefinitionView{Definition: rollout.Plan})
	if err != nil {
		return err
	}

	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rErr != nil {
			_ = rollback()
			return
		}
		rErr = commit()
	}()

	if res, err := tx.ExecContext(ctx, releaseRolloutUpdateQuery,
		rollout.ID,
		rollout.FlagID,
		rollout.DeploymentEnvironmentID,
		planJSON,
	); err != nil {
		return err
	} else {
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return fmt.Errorf(`release rollout not found`)
		}
	}

	return pg.notify(ctx, tx, releaseRolloutUpdateSubscriptionName, *rollout)
}

const pilotUpdateQuery = `
UPDATE "release_pilots"
SET flag_id = $2,
	env_id = $3,
    external_id = $4,
    is_participating = $5
WHERE id = $1;
`

func (pg *Postgres) releaseManualPilotUpdate(ctx context.Context, pilot *release.ManualPilot) (rErr error) {
	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rErr != nil {
			_ = rollback()
			return
		}
		rErr = commit()
	}()

	if res, err := tx.ExecContext(ctx, pilotUpdateQuery, pilot.ID,
		pilot.FlagID,
		pilot.DeploymentEnvironmentID,
		pilot.ExternalID,
		pilot.IsParticipating,
	); err != nil {
		return err
	} else {
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return fmt.Errorf(`release pilot not found`)
		}
	}

	return pg.notify(ctx, tx, releaseManualPilotUpdateSubscriptionName, *pilot)
}

const tokenUpdateQuery = `
UPDATE "tokens"
SET sha512 = $1,
    owner_uid = $2,
    issued_at = $3,
    duration = $4
WHERE id = $5;
`

func (pg *Postgres) securityTokenUpdate(ctx context.Context, t *security.Token) (rErr error) {
	tx, commit, rollback, err := pg.getTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rErr != nil {
			_ = rollback()
			return
		}
		rErr = commit()
	}()

	res, err := tx.ExecContext(ctx, tokenUpdateQuery,
		t.SHA512,
		t.OwnerUID,
		t.IssuedAt,
		t.Duration,
		t.ID,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return fmt.Errorf(`security token not found`)
	}

	return pg.notify(ctx, tx, securityTokenUpdateSubscriptionName, *t)
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
		const err consterror.Error = "Type assertion .([]byte) failed."
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

//--------------------------------------------------------------------------------------------------------------------//

func (pg *Postgres) notify(ctx context.Context, tx *sql.Tx, name string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `SELECT pg_notify($1, $2)`, name, string(data))
	return err
}

func newPostgresSubscription(connstr string, name string, T interface{}, subscriber resources.Subscriber) (*postgresSubscription, error) {
	const (
		minReconnectInterval = 10 * time.Second
		maxReconnectInterval = time.Minute
	)
	var sub postgresSubscription
	sub.rType = reflect.TypeOf(T)
	sub.subscriber = subscriber
	sub.listener = pq.NewListener(connstr, minReconnectInterval, maxReconnectInterval, sub.reportProblemToSubscriber)
	sub.exit.context, sub.exit.signaler = context.WithCancel(context.Background())
	return &sub, sub.start(name)
}

type postgresSubscription struct {
	// T       interface{}
	rType      reflect.Type
	subscriber resources.Subscriber
	listener   *pq.Listener
	exit       struct {
		wg       sync.WaitGroup
		context  context.Context
		signaler func()
	}
}

func (sub *postgresSubscription) start(name string) error {
	if err := sub.listener.Listen(name); err != nil {
		return err
	}

	sub.exit.wg.Add(1)
	go sub.handler()
	return nil
}

func (sub *postgresSubscription) handler() {
	defer sub.exit.wg.Done()

wrk:
	for {
		ctx := context.Background()

		select {
		case <-sub.exit.context.Done():
			break wrk

		case n := <-sub.listener.Notify:
			ptr := reflect.New(sub.rType)

			if sub.handleError(ctx, json.Unmarshal([]byte(n.Extra), ptr.Interface())) {
				continue wrk
			}

			sub.handleError(ctx, sub.subscriber.Handle(ctx, ptr.Elem().Interface()))

			continue wrk
		case <-time.After(time.Minute):
			sub.handleError(ctx, sub.listener.Ping())
			continue wrk
		}
	}
}

func (sub *postgresSubscription) handleError(ctx context.Context, err error) (isErrorHandled bool) {
	if err == nil {
		return false
	}

	if sErr := sub.subscriber.Error(ctx, err); sErr != nil {
		log.Println(`ERROR`, sErr.Error())
	}

	return true
}

func (sub *postgresSubscription) Close() error {
	if sub.exit.signaler == nil || sub.listener == nil {
		return nil
	}

	sub.exit.signaler()
	sub.exit.wg.Wait()
	return sub.listener.Close()
}

func (sub *postgresSubscription) reportProblemToSubscriber(ev pq.ListenerEventType, err error) {
	if err != nil {
		_ = sub.subscriber.Error(context.Background(), err)
	}
}

const (
	releaseFlagCreateSubscriptionName     = `create_release_flag`
	releaseFlagUpdateSubscriptionName     = `update_release_flag`
	releaseFlagDeleteByIDSubscriptionName = `delete_by_id_release_flag`
	releaseFlagDeleteAllSubscriptionName  = `delete_all_release_flag`

	releaseManualPilotCreateSubscriptionName     = `create_release_manual_pilot`
	releaseManualPilotUpdateSubscriptionName     = `update_release_manual_pilot`
	releaseManualPilotDeleteByIDSubscriptionName = `delete_by_id_release_manual_pilot`
	releaseManualPilotDeleteAllSubscriptionName  = `delete_all_release_manual_pilot`

	releaseRolloutCreateSubscriptionName     = `create_release_rollout`
	releaseRolloutUpdateSubscriptionName     = `update_release_rollout`
	releaseRolloutDeleteByIDSubscriptionName = `delete_by_id_release_rollout`
	releaseRolloutDeleteAllSubscriptionName  = `delete_all_release_rollout`

	securityTokenCreateSubscriptionName     = `create_security_token`
	securityTokenUpdateSubscriptionName     = `update_security_token`
	securityTokenDeleteByIDSubscriptionName = `delete_by_id_security_token`
	securityTokenDeleteAllSubscriptionName  = `delete_all_security_token`

	deploymentEnvironmentCreateSubscriptionName     = `create_deployment_environment`
	deploymentEnvironmentUpdateSubscriptionName     = `update_deployment_environment`
	deploymentEnvironmentDeleteByIDSubscriptionName = `delete_by_id_deployment_environment`
	deploymentEnvironmentDeleteAllSubscriptionName  = `delete_all_deployment_environment`
)

// no support for tx
func (pg *Postgres) SubscribeToCreate(ctx context.Context, T interface{}, subscriber resources.Subscriber) (resources.Subscription, error) {
	switch T.(type) {
	case release.Flag:
		return newPostgresSubscription(pg.DSN, releaseFlagCreateSubscriptionName, T, subscriber)
	case release.ManualPilot:
		return newPostgresSubscription(pg.DSN, releaseManualPilotCreateSubscriptionName, T, subscriber)
	case release.Rollout:
		return newPostgresSubscription(pg.DSN, releaseRolloutCreateSubscriptionName, T, subscriber)
	case security.Token:
		return newPostgresSubscription(pg.DSN, securityTokenCreateSubscriptionName, T, subscriber)
	case deployment.Environment:
		return newPostgresSubscription(pg.DSN, deploymentEnvironmentCreateSubscriptionName, T, subscriber)
	default:
		return nil, fmt.Errorf(`ErrNotImplemented`)
	}
}

func (pg *Postgres) SubscribeToUpdate(ctx context.Context, T resources.T, subscriber resources.Subscriber) (resources.Subscription, error) {
	switch T.(type) {
	case release.Flag:
		return newPostgresSubscription(pg.DSN, releaseFlagUpdateSubscriptionName, T, subscriber)
	case release.ManualPilot:
		return newPostgresSubscription(pg.DSN, releaseManualPilotUpdateSubscriptionName, T, subscriber)
	case release.Rollout:
		return newPostgresSubscription(pg.DSN, releaseRolloutUpdateSubscriptionName, T, subscriber)
	case security.Token:
		return newPostgresSubscription(pg.DSN, securityTokenUpdateSubscriptionName, T, subscriber)
	case deployment.Environment:
		return newPostgresSubscription(pg.DSN, deploymentEnvironmentUpdateSubscriptionName, T, subscriber)
	default:
		return nil, fmt.Errorf(`ErrNotImplemented`)
	}
}

func (pg *Postgres) SubscribeToDeleteByID(ctx context.Context, T resources.T, subscriber resources.Subscriber) (resources.Subscription, error) {
	switch T.(type) {
	case release.Flag:
		return newPostgresSubscription(pg.DSN, releaseFlagDeleteByIDSubscriptionName, T, subscriber)
	case release.ManualPilot:
		return newPostgresSubscription(pg.DSN, releaseManualPilotDeleteByIDSubscriptionName, T, subscriber)
	case release.Rollout:
		return newPostgresSubscription(pg.DSN, releaseRolloutDeleteByIDSubscriptionName, T, subscriber)
	case security.Token:
		return newPostgresSubscription(pg.DSN, securityTokenDeleteByIDSubscriptionName, T, subscriber)
	case deployment.Environment:
		return newPostgresSubscription(pg.DSN, deploymentEnvironmentDeleteByIDSubscriptionName, T, subscriber)
	default:
		return nil, fmt.Errorf(`ErrNotImplemented`)
	}
}

func (pg *Postgres) SubscribeToDeleteAll(ctx context.Context, T resources.T, subscriber resources.Subscriber) (resources.Subscription, error) {
	switch T.(type) {
	case release.Flag:
		return newPostgresSubscription(pg.DSN, releaseFlagDeleteAllSubscriptionName, T, subscriber)
	case release.ManualPilot:
		return newPostgresSubscription(pg.DSN, releaseManualPilotDeleteAllSubscriptionName, T, subscriber)
	case release.Rollout:
		return newPostgresSubscription(pg.DSN, releaseRolloutDeleteAllSubscriptionName, T, subscriber)
	case security.Token:
		return newPostgresSubscription(pg.DSN, securityTokenDeleteAllSubscriptionName, T, subscriber)
	case deployment.Environment:
		return newPostgresSubscription(pg.DSN, deploymentEnvironmentDeleteAllSubscriptionName, T, subscriber)
	default:
		return nil, fmt.Errorf(`ErrNotImplemented`)
	}
}
