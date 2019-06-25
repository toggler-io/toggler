package postgres

import (
	"database/sql"
	"fmt"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/frameless/resources/storages/memorystorage"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
	"net/url"
	"strconv"
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

func (pg *Postgres) Save(entity interface{}) error {
	if currentID, ok := specs.LookupID(entity); !ok || currentID != "" {
		return fmt.Errorf("entity already have an ID: %s", currentID)
	}

	switch e := entity.(type) {
	case *rollouts.FeatureFlag:
		return pg.featureFlagInsertNew(e)
	case *rollouts.Pilot:
		return pg.pilotInsertNew(e)
	case *security.Token:
		return pg.tokenInsertNew(e)
	case *specs.TestEntity:
		return pg.testEntities.Save(entity)
	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) FindByID(ID string, ptr interface{}) (bool, error) {
	switch ptr.(type) {
	case *specs.TestEntity:
		return pg.testEntities.FindByID(ID, ptr)
	}

	id, err := strconv.ParseInt(ID, 10, 64)

	if err != nil {
		return false, nil
	}

	switch e := ptr.(type) {
	case *rollouts.FeatureFlag:
		return pg.featureFlagFindByID(id, e)

	case *rollouts.Pilot:
		return pg.pilotFindByID(id, e)

	case *security.Token:
		return pg.tokenFindByID(id, e)

	default:
		return false, frameless.ErrNotImplemented
	}
}

func (pg *Postgres) Truncate(Type interface{}) error {
	switch Type.(type) {
	case specs.TestEntity, *specs.TestEntity:
		return pg.testEntities.Truncate(Type)
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

	_, err := pg.DB.Exec(fmt.Sprintf(`TRUNCATE TABLE "%s"`, tableName))
	return err
}

func (pg *Postgres) DeleteByID(Type interface{}, ID string) error {
	switch Type.(type) {
	case specs.TestEntity, *specs.TestEntity:
		return pg.testEntities.DeleteByID(Type, ID)
	}

	id, err := strconv.ParseInt(ID, 10, 64)

	if err != nil {
		return err
	}

	switch Type.(type) {
	case rollouts.FeatureFlag, *rollouts.FeatureFlag:
		_, err := pg.DB.Exec(`DELETE FROM "feature_flags" WHERE id = $1`, id)
		return err

	case rollouts.Pilot, *rollouts.Pilot:
		_, err := pg.DB.Exec(`DELETE FROM "pilots" WHERE id = $1`, id)
		return err

	case security.Token, *security.Token:
		_, err := pg.DB.Exec(`DELETE FROM "tokens" WHERE id = $1`, id)
		return err

	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) Update(ptr interface{}) error {
	switch e := ptr.(type) {
	case *rollouts.FeatureFlag:
		return pg.featureFlagUpdate(e)

	case *rollouts.Pilot:
		return pg.pilotUpdate(e)

	case *security.Token:
		return pg.tokenUpdate(e)

	case *specs.TestEntity:
		return pg.testEntities.Update(ptr)

	default:
		return frameless.ErrNotImplemented
	}
}

func (pg *Postgres) FindAll(Type interface{}) frameless.Iterator {
	switch Type.(type) {
	case rollouts.FeatureFlag, *rollouts.FeatureFlag:
		return pg.featureFlagFindAll()

	case rollouts.Pilot, *rollouts.Pilot:
		return pg.pilotFindAll()

	case security.Token, *security.Token:
		return pg.tokenFindAll()

	case specs.TestEntity, *specs.TestEntity:
		return pg.FindAll(Type)

	default:
		return iterators.NewError(frameless.ErrNotImplemented)
	}
}

const featureFlagFindByNameQuery = `
SELECT id, name, rollout_rand_seed, rollout_strategy_percentage, rollout_strategy_decision_logic_api 
FROM feature_flags 
WHERE name = $1;
`

func (pg *Postgres) FindFlagByName(name string) (*rollouts.FeatureFlag, error) {

	row := pg.DB.QueryRow(featureFlagFindByNameQuery, name)

	var ff rollouts.FeatureFlag
	var DecisionLogicAPI sql.NullString

	err := row.Scan(
		&ff.ID,
		&ff.Name,
		&ff.Rollout.RandSeed,
		&ff.Rollout.Strategy.Percentage,
		&DecisionLogicAPI,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if DecisionLogicAPI.Valid {

		u, err := url.ParseRequestURI(DecisionLogicAPI.String)

		if err != nil {
			return nil, err
		}

		ff.Rollout.Strategy.DecisionLogicAPI = u

	}

	return &ff, nil

}

const findFlagPilotByExternalPilotIDQuery = `
SELECT id, feature_flag_id, external_id, enrolled 
FROM "pilots"
WHERE feature_flag_id = $1
  AND external_id = $2
`

func (pg *Postgres) FindFlagPilotByExternalPilotID(FeatureFlagID, ExternalPilotID string) (*rollouts.Pilot, error) {
	flagID, err := strconv.ParseInt(FeatureFlagID, 10, 64)

	if err != nil {
		return nil, nil
	}

	row := pg.DB.QueryRow(findFlagPilotByExternalPilotIDQuery, flagID, ExternalPilotID)

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

func (pg *Postgres) FindPilotsByFeatureFlag(ff *rollouts.FeatureFlag) frameless.Iterator {

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

	rows, err := pg.DB.Query(findPilotsByFeatureFlagQuery, ffID)

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
SELECT id, token, duration, issued_at, owner_uid
FROM "tokens" 
WHERE token = $1;
`

func (pg *Postgres) FindTokenByTokenString(token string) (*security.Token, error) {
	row := pg.DB.QueryRow(tokenFindByTokenStringQuery, token)
	var t security.Token

	err := row.Scan(
		&t.ID,
		&t.Token,
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

func (pg *Postgres) featureFlagInsertNew(flag *rollouts.FeatureFlag) error {

	var DecisionLogicAPI sql.NullString

	if flag.Rollout.Strategy.DecisionLogicAPI != nil {
		DecisionLogicAPI.Valid = true
		DecisionLogicAPI.String = flag.Rollout.Strategy.DecisionLogicAPI.String()
	}

	row := pg.DB.QueryRow(featureFlagInsertNewQuery,
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

func (pg *Postgres) pilotInsertNew(pilot *rollouts.Pilot) error {

	flagID, err := strconv.ParseInt(pilot.FeatureFlagID, 10, 64)

	if err != nil {
		return fmt.Errorf(`invalid Feature Flag ID: ` + pilot.FeatureFlagID)
	}

	row := pg.DB.QueryRow(pilotInsertNewQuery,
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
INSERT INTO "tokens" (token, owner_uid, issued_at, duration)
VALUES ($1, $2, $3, $4)
RETURNING id;
`

func (pg *Postgres) tokenInsertNew(token *security.Token) error {
	row := pg.DB.QueryRow(tokenInsertNewQuery,
		token.Token,
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

const featureFlagFindByIDQuery = `
SELECT id, name, rollout_rand_seed, rollout_strategy_percentage, rollout_strategy_decision_logic_api 
FROM feature_flags 
WHERE id = $1;
`

func (pg *Postgres) featureFlagFindByID(id int64, flag *rollouts.FeatureFlag) (bool, error) {
	row := pg.DB.QueryRow(featureFlagFindByIDQuery, id)
	var ff rollouts.FeatureFlag

	var DecisionLogicAPI sql.NullString

	err := row.Scan(
		&ff.ID,
		&ff.Name,
		&ff.Rollout.RandSeed,
		&ff.Rollout.Strategy.Percentage,
		&DecisionLogicAPI,
	)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if DecisionLogicAPI.Valid {

		u, err := url.ParseRequestURI(DecisionLogicAPI.String)

		if err != nil {
			return false, err
		}

		ff.Rollout.Strategy.DecisionLogicAPI = u

	}

	*flag = ff
	return true, nil
}

const pilotFindByIDQuery = `
SELECT id, feature_flag_id, external_id, enrolled 
FROM "pilots" 
WHERE id = $1;
`

func (pg *Postgres) pilotFindByID(id int64, pilot *rollouts.Pilot) (bool, error) {
	row := pg.DB.QueryRow(pilotFindByIDQuery, id)
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
SELECT id, token, duration, issued_at, owner_uid
FROM "tokens" 
WHERE id = $1;
`

func (pg *Postgres) tokenFindByID(id int64, token *security.Token) (bool, error) {
	row := pg.DB.QueryRow(tokenFindByIDQuery, id)
	var t security.Token

	err := row.Scan(
		&t.ID,
		&t.Token,
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

func (pg *Postgres) featureFlagFindAll() frameless.Iterator {
	rows, err := pg.DB.Query(featureFlagFindAllQuery)

	if err != nil {
		return iterators.NewError(err)
	}

	receiver, sender := iterators.NewPipe()

	go func() {
		defer sender.Close()

	wrk:
		for rows.Next() {

			var ff rollouts.FeatureFlag

			var DecisionLogicAPI sql.NullString

			err := rows.Scan(
				&ff.ID,
				&ff.Name,
				&ff.Rollout.RandSeed,
				&ff.Rollout.Strategy.Percentage,
				&DecisionLogicAPI,
			)

			if err == sql.ErrNoRows {
				break wrk
			}

			if err != nil {
				sender.Error(err)
				break wrk
			}

			if DecisionLogicAPI.Valid {

				u, err := url.ParseRequestURI(DecisionLogicAPI.String)

				if err != nil {
					sender.Error(err)
					break wrk
				}

				ff.Rollout.Strategy.DecisionLogicAPI = u

			}

			if err := sender.Encode(&ff); err != nil {
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

const pilotFindAllQuery = `
SELECT id, feature_flag_id, external_id, enrolled 
FROM "pilots"
`

func (pg *Postgres) pilotFindAll() frameless.Iterator {
	rows, err := pg.DB.Query(pilotFindAllQuery)

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
SELECT id, token, duration, issued_at, owner_uid
FROM "tokens"
`

func (pg *Postgres) tokenFindAll() frameless.Iterator {
	rows, err := pg.DB.Query(tokenFindAllQuery)

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
				&t.Token,
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

func (pg *Postgres) featureFlagUpdate(flag *rollouts.FeatureFlag) error {
	var DecisionLogicAPI sql.NullString

	if flag.Rollout.Strategy.DecisionLogicAPI != nil {
		DecisionLogicAPI.Valid = true
		DecisionLogicAPI.String = flag.Rollout.Strategy.DecisionLogicAPI.String()
	}

	_, err := pg.DB.Exec(featureFlagUpdateQuery,
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

func (pg *Postgres) pilotUpdate(pilot *rollouts.Pilot) error {
	_, err := pg.DB.Exec(pilotUpdateQuery,
		pilot.FeatureFlagID,
		pilot.ExternalID,
		pilot.Enrolled,
		pilot.ID,
	)

	return err
}

const tokenUpdateQuery = `
UPDATE "tokens"
SET token = $1,
    owner_uid = $2,
    issued_at = $3,
    duration = $4
WHERE id = $5;
`

func (pg *Postgres) tokenUpdate(t *security.Token) error {
	_, err := pg.DB.Exec(tokenUpdateQuery,
		t.Token,
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
