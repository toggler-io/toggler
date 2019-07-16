package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/frameless/resources/storages/memorystorage"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
	_ "github.com/lib/pq"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func NewPostgres(db *sql.DB) (*Postgres, error) {
	pg := &Postgres{DB: db, testEntities: memorystorage.NewMemory()}

	if err := pg.Migrate(); err != nil {
		return nil, err
	}

	return pg, nil
}

type Postgres struct {
	*sql.DB
	timeLocation *time.Location
	testEntities *memorystorage.Memory
}

func (pg *Postgres) Save(ctx context.Context, entity interface{}) error {
	if currentID, ok := specs.LookupID(entity); !ok || currentID != "" {
		return fmt.Errorf("entity already have an ID: %s", currentID)
	}

	switch e := entity.(type) {
	case *rollouts.FeatureFlag:
		return pg.featureFlagInsertNew(ctx, e)
	case *rollouts.Pilot:
		return pg.pilotInsertNew(ctx, e)
	case *security.Token:
		return pg.tokenInsertNew(ctx, e)
	case *specs.TestEntity:
		return pg.testEntities.Save(ctx, entity)
	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) FindByID(ctx context.Context, ptr interface{}, ID string) (bool, error) {
	switch ptr.(type) {
	case *specs.TestEntity:
		return pg.testEntities.FindByID(ctx, ptr, ID)
	}

	id, err := strconv.ParseInt(ID, 10, 64)

	if err != nil {
		return false, nil
	}

	switch e := ptr.(type) {
	case *rollouts.FeatureFlag:
		return pg.featureFlagFindByID(ctx, e, id)

	case *rollouts.Pilot:
		return pg.pilotFindByID(ctx, e, id)

	case *security.Token:
		return pg.tokenFindByID(ctx, e, id)

	default:
		return false, frameless.ErrNotImplemented
	}
}

func (pg *Postgres) Truncate(ctx context.Context, Type interface{}) error {
	switch Type.(type) {
	case specs.TestEntity, *specs.TestEntity:
		return pg.testEntities.Truncate(ctx, Type)
	}

	var tableName string
	switch Type.(type) {
	case rollouts.FeatureFlag, *rollouts.FeatureFlag:
		tableName = `feature_flags`
	case rollouts.Pilot, *rollouts.Pilot:
		tableName = `pilots`
	case security.Token, *security.Token:
		tableName = `tokens`
	default:
		return frameless.ErrNotImplemented
	}

	_, err := pg.DB.ExecContext(ctx, fmt.Sprintf(`TRUNCATE TABLE "%s"`, tableName))
	return err
}

func (pg *Postgres) DeleteByID(ctx context.Context, Type interface{}, ID string) error {
	switch Type.(type) {
	case specs.TestEntity, *specs.TestEntity:
		return pg.testEntities.DeleteByID(ctx, Type, ID)
	}

	id, err := strconv.ParseInt(ID, 10, 64)

	if err != nil {
		return err
	}

	switch Type.(type) {
	case rollouts.FeatureFlag, *rollouts.FeatureFlag:
		_, err := pg.DB.ExecContext(ctx, `DELETE FROM "feature_flags" WHERE id = $1`, id)
		return err

	case rollouts.Pilot, *rollouts.Pilot:
		_, err := pg.DB.ExecContext(ctx, `DELETE FROM "pilots" WHERE id = $1`, id)
		return err

	case security.Token, *security.Token:
		_, err := pg.DB.ExecContext(ctx, `DELETE FROM "tokens" WHERE id = $1`, id)
		return err

	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) Update(ctx context.Context, ptr interface{}) error {
	switch e := ptr.(type) {
	case *rollouts.FeatureFlag:
		return pg.featureFlagUpdate(ctx, e)

	case *rollouts.Pilot:
		return pg.pilotUpdate(ctx, e)

	case *security.Token:
		return pg.tokenUpdate(ctx, e)

	case *specs.TestEntity:
		return pg.testEntities.Update(ctx, ptr)

	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) FindAll(ctx context.Context, Type interface{}) frameless.Iterator {
	switch Type.(type) {
	case rollouts.FeatureFlag, *rollouts.FeatureFlag:
		return pg.featureFlagFindAll(ctx)

	case rollouts.Pilot, *rollouts.Pilot:
		return pg.pilotFindAll(ctx)

	case security.Token, *security.Token:
		return pg.tokenFindAll(ctx)

	case specs.TestEntity, *specs.TestEntity:
		return pg.FindAll(ctx, Type)

	default:
		return iterators.NewError(frameless.ErrNotImplemented)
	}
}

func (pg *Postgres) FindFlagByName(ctx context.Context, name string) (*rollouts.FeatureFlag, error) {

	mapper := featureFlagMapper{}
	query := fmt.Sprintf(`%s FROM "feature_flags" WHERE "name" = $1`,
		mapper.SelectClause())

	row := pg.DB.QueryRowContext(ctx, query, name)

	var ff rollouts.FeatureFlag

	err := mapper.Map(row, &ff)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &ff, nil

}

const findFlagPilotByExternalPilotIDQuery = `
SELECT id, feature_flag_id, external_id, enrolled 
FROM "pilots"
WHERE feature_flag_id = $1
  AND external_id = $2
`

func (pg *Postgres) FindFlagPilotByExternalPilotID(ctx context.Context, FeatureFlagID, ExternalPilotID string) (*rollouts.Pilot, error) {
	flagID, err := strconv.ParseInt(FeatureFlagID, 10, 64)

	if err != nil {
		return nil, nil
	}

	row := pg.DB.QueryRowContext(ctx, findFlagPilotByExternalPilotIDQuery, flagID, ExternalPilotID)

	var p rollouts.Pilot

	err = row.Scan(
		&p.ID,
		&p.FeatureFlagID,
		&p.ExternalID,
		&p.Enrolled,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}

const findPilotsByFeatureFlagQuery = `
SELECT id, feature_flag_id, external_id, enrolled 
FROM "pilots"
WHERE feature_flag_id = $1
`

func (pg *Postgres) FindPilotsByFeatureFlag(ctx context.Context, ff *rollouts.FeatureFlag) frameless.Iterator {

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

	rows, err := pg.DB.QueryContext(ctx, findPilotsByFeatureFlagQuery, ffID)

	if err != nil {
		return iterators.NewError(err)
	}

	receiver, sender := iterators.NewPipe()

	go func() {
		defer sender.Close()

	wrk:
		for rows.Next() {

			var p rollouts.Pilot

			err := rows.Scan(
				&p.ID,
				&p.FeatureFlagID,
				&p.ExternalID,
				&p.Enrolled,
			)

			if err == sql.ErrNoRows {
				break wrk
			}

			if err != nil {
				sender.Error(err)
				break wrk
			}

			if err := sender.Encode(&p); err != nil {
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

func (pg *Postgres) FindFlagsByName(ctx context.Context, flagNames ...string) frameless.Iterator {

	var namesInClause []string
	var args []interface{}

	namesInClause = append(namesInClause)
	for i, arg := range flagNames {
		namesInClause = append(namesInClause, fmt.Sprintf(`$%d`, i+1))
		args = append(args, arg)
	}

	mapper := featureFlagMapper{}

	query := fmt.Sprintf(`%s FROM "feature_flags" WHERE "name" IN (%s)`,
		mapper.SelectClause(),
		strings.Join(namesInClause, `,`))

	flags, err := pg.DB.QueryContext(ctx, query, args...)

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(flags, mapper)
}

const featureFlagInsertNewQuery = `
INSERT INTO "feature_flags" (
	name,
	rollout_rand_seed,
	rollout_strategy_percentage,
	rollout_strategy_decision_logic_api
)
VALUES ($1, $2, $3, $4)
RETURNING id;
`

func (pg *Postgres) featureFlagInsertNew(ctx context.Context, flag *rollouts.FeatureFlag) error {

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

	return specs.SetID(flag, strconv.FormatInt(id.Int64, 10))
}

const pilotInsertNewQuery = `
INSERT INTO "pilots" (feature_flag_id, external_id, enrolled)
VALUES ($1, $2, $3)
RETURNING id;
`

func (pg *Postgres) pilotInsertNew(ctx context.Context, pilot *rollouts.Pilot) error {

	flagID, err := strconv.ParseInt(pilot.FeatureFlagID, 10, 64)

	if err != nil {
		return fmt.Errorf(`invalid Feature Flag ID: ` + pilot.FeatureFlagID)
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

	return specs.SetID(pilot, strconv.FormatInt(id.Int64, 10))
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

	return specs.SetID(token, strconv.FormatInt(id.Int64, 10))
}

func (pg *Postgres) featureFlagFindByID(ctx context.Context, flag *rollouts.FeatureFlag, id int64) (bool, error) {

	mapper := featureFlagMapper{}
	query := fmt.Sprintf(`%s FROM "feature_flags" WHERE "id" = $1`, mapper.SelectClause())
	row := pg.DB.QueryRowContext(ctx, query, id)

	var ff rollouts.FeatureFlag
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

const pilotFindByIDQuery = `
SELECT id, feature_flag_id, external_id, enrolled 
FROM "pilots" 
WHERE id = $1;
`

func (pg *Postgres) pilotFindByID(ctx context.Context, pilot *rollouts.Pilot, id int64) (bool, error) {
	row := pg.DB.QueryRowContext(ctx, pilotFindByIDQuery, id)
	var p rollouts.Pilot

	err := row.Scan(
		&p.ID,
		&p.FeatureFlagID,
		&p.ExternalID,
		&p.Enrolled,
	)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	*pilot = p
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

const featureFlagFindAllQuery = `
SELECT id, name, rollout_rand_seed, rollout_strategy_percentage, rollout_strategy_decision_logic_api 
FROM feature_flags 
`

func (pg *Postgres) featureFlagFindAll(ctx context.Context) frameless.Iterator {
	mapper := featureFlagMapper{}
	query := fmt.Sprintf(`%s FROM "feature_flags"`, mapper.SelectClause())
	rows, err := pg.DB.QueryContext(ctx, query)
	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, mapper)
}

const pilotFindAllQuery = `
SELECT id, feature_flag_id, external_id, enrolled 
FROM "pilots"
`

func (pg *Postgres) pilotFindAll(ctx context.Context) frameless.Iterator {
	rows, err := pg.DB.QueryContext(ctx, pilotFindAllQuery)

	if err != nil {
		return iterators.NewError(err)
	}

	receiver, sender := iterators.NewPipe()

	go func() {
		defer sender.Close()

	wrk:
		for rows.Next() {

			var p rollouts.Pilot

			err := rows.Scan(
				&p.ID,
				&p.FeatureFlagID,
				&p.ExternalID,
				&p.Enrolled,
			)

			if err == sql.ErrNoRows {
				break wrk
			}

			if err != nil {
				sender.Error(err)
				break wrk
			}

			if err := sender.Encode(&p); err != nil {
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
UPDATE "feature_flags"
SET name = $1,
    rollout_rand_seed = $2,
    rollout_strategy_percentage = $3,
    rollout_strategy_decision_logic_api = $4
WHERE id = $5;
`

func (pg *Postgres) featureFlagUpdate(ctx context.Context, flag *rollouts.FeatureFlag) error {
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

func (pg *Postgres) pilotUpdate(ctx context.Context, pilot *rollouts.Pilot) error {
	_, err := pg.DB.ExecContext(ctx, pilotUpdateQuery,
		pilot.FeatureFlagID,
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
	var ff rollouts.FeatureFlag
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
