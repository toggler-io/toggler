package storages

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/lazyloading"
	"github.com/adamluzsi/frameless/postgresql"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/lib/pq"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"

	"github.com/toggler-io/toggler/external/resource/storages/migrations"
)

func NewPostgres(dsn string) (*Postgres, error) {
	cm := postgresql.NewConnectionManager(dsn)

	// TODO: move out migration from initialization
	if err := migrations.MigratePostgres(dsn); err != nil {
		return nil, err
	}

	postgres := &Postgres{
		DSN:               dsn,
		ConnectionManager: cm,
	}
	return postgres, postgres.Init()
}

type Postgres struct {
	DSN string
	postgresql.ConnectionManager

	init sync.Once

	subs struct {
		listener  *pq.Listener
		lock      sync.Mutex
		callbacks []pq.EventCallbackType
	}
	storage struct {
		ReleaseFlag        lazyloading.Var
		ReleasePilot       lazyloading.Var
		ReleaseRollout     lazyloading.Var
		ReleaseEnvironment lazyloading.Var
		SecurityToken      lazyloading.Var
	}
}

func (p *Postgres) Init() (rErr error) {
	p.init.Do(func() {
		// toggler#Storage
		// releases#Storage
		// FlagStorage
		// Publisher
		// CreatorPublisher
		// describe .Subscribe/Create
		// and events made
		// and then new subscriberGet registered
		// and further events made
		// then new subscriberGet will receive new events
		// and it fails because in some case old events are not received when a new subscriber is made.
		//
		//const reconnectMinInterval = 10 * time.Second
		//const reconnectMaxInterval = time.Minute
		//p.subs.listener = pq.NewListener(p.DSN, reconnectMinInterval, reconnectMaxInterval, func(event pq.ListenerEventType, err error) {
		//	for _, cb := range p.subs.callbacks {
		//		cb(event, err)
		//	}
		//})
	})
	return
}

func (p *Postgres) Close() error {
	if p.ConnectionManager == nil {
		return nil
	}
	return p.ConnectionManager.Close()
}

func (p *Postgres) mkPostgresqlStorage(T interface{}, m postgresql.Mapping) *postgresql.Storage {
	listenNotifySM := postgresql.NewListenNotifySubscriptionManager(T, m, p.DSN, p.ConnectionManager)
	//listenNotifySM.Listener = p.subs.listener
	//p.subs.lock.Lock()
	//p.subs.callbacks = append(p.subs.callbacks, listenNotifySM.ListenerEventCallback)
	//p.subs.lock.Unlock()
	return &postgresql.Storage{
		T:                   T,
		Mapping:             m,
		ConnectionManager:   p.ConnectionManager,
		SubscriptionManager: listenNotifySM,
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *Postgres) ReleaseFlag(ctx context.Context) release.FlagStorage {
	return p.storage.ReleaseFlag.Do(func() interface{} {
		return ReleaseFlagPgStorage{
			Storage: p.mkPostgresqlStorage(release.Flag{},
				postgresql.Mapper{
					Table:   "release_flags",
					ID:      "id",
					NewIDFn: newIDFn,
					Columns: []string{"id", "name"},
					ToArgsFn: func(ptr interface{}) ([]interface{}, error) {
						e := ptr.(*release.Flag)
						return []interface{}{e.ID, e.Name}, nil
					},
					MapFn: func(s iterators.SQLRowScanner, ptr interface{}) error {
						e := ptr.(*release.Flag)
						return s.Scan(&e.ID, &e.Name)
					},
				}),
		}
	}).(ReleaseFlagPgStorage)
}

type ReleaseFlagPgStorage struct {
	*postgresql.Storage
}

func (s ReleaseFlagPgStorage) FindByName(ctx context.Context, name string) (*release.Flag, error) {
	m := s.Mapping
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE "name" = $1`, toSelectClause(m), m.TableRef())

	c, err := s.ConnectionManager.Connection(ctx)
	if err != nil {
		return nil, err
	}

	row := c.QueryRowContext(ctx, query, name)
	var ff release.Flag
	err = m.Map(row, &ff)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ff, nil
}

func (s ReleaseFlagPgStorage) FindByNames(ctx context.Context, names ...string) release.FlagEntries {
	var namesInClause []string
	var args []interface{}

	namesInClause = append(namesInClause)
	for i, arg := range names {
		namesInClause = append(namesInClause, fmt.Sprintf(`$%d`, i+1))
		args = append(args, arg)
	}

	m := s.Mapping

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE "name" IN (%s)`,
		toSelectClause(m),
		m.TableRef(),
		strings.Join(namesInClause, `,`))

	c, err := s.Storage.ConnectionManager.Connection(ctx)
	if err != nil {
		return iterators.NewError(err)
	}

	flags, err := c.QueryContext(ctx, query, args...)

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(flags, m)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *Postgres) ReleasePilot(ctx context.Context) release.PilotStorage {
	return p.storage.ReleasePilot.Do(func() interface{} {
		return ReleasePilotPgStorage{
			Storage: p.mkPostgresqlStorage(release.Pilot{}, postgresql.Mapper{
				Table:   "release_pilots",
				ID:      "id",
				NewIDFn: newIDFn,
				Columns: []string{
					`id`,
					`flag_id`,
					`env_id`,
					`public_id`,
					`is_participating`,
				},
				ToArgsFn: func(ptr interface{}) ([]interface{}, error) {
					e := ptr.(*release.Pilot)
					return []interface{}{
						e.ID,
						e.FlagID,
						e.EnvironmentID,
						e.PublicID,
						e.IsParticipating,
					}, nil
				},
				MapFn: func(s iterators.SQLRowScanner, ptr interface{}) error {
					e := ptr.(*release.Pilot)
					return s.Scan(
						&e.ID,
						&e.FlagID,
						&e.EnvironmentID,
						&e.PublicID,
						&e.IsParticipating,
					)
				},
			}),
		}
	}).(ReleasePilotPgStorage)
}

type ReleasePilotPgStorage struct {
	*postgresql.Storage
}

func (s ReleasePilotPgStorage) FindByFlagEnvPublicID(ctx context.Context, flagID, envID interface{}, pilotExtID string) (*release.Pilot, error) {
	if !isUUIDValid(flagID) {
		return nil, nil
	}

	m := s.Mapping
	q := fmt.Sprintf(`SELECT %s FROM %s WHERE "flag_id" = $1 AND "env_id" = $2 AND "public_id" = $3`,
		toSelectClause(m),
		m.TableRef(),
	)

	c, err := s.ConnectionManager.Connection(ctx)
	if err != nil {
		return nil, err
	}

	row := c.QueryRowContext(ctx, q, flagID, envID, pilotExtID)

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

func (s ReleasePilotPgStorage) FindByFlag(ctx context.Context, flag release.Flag) release.PilotEntries {
	if flag.ID == `` {
		return iterators.NewEmpty()
	}

	if flag.ID == `` {
		return iterators.NewEmpty()
	}

	if !isUUIDValid(flag.ID) {
		return iterators.NewEmpty()
	}

	m := s.Mapping
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE "flag_id" = $1`, toSelectClause(m), m.TableRef())

	c, err := s.ConnectionManager.Connection(ctx)
	if err != nil {
		return iterators.NewError(err)
	}

	rows, err := c.QueryContext(ctx, query, flag.ID)

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSQLRows(rows, m)
}

func (s ReleasePilotPgStorage) FindByPublicID(ctx context.Context, externalID string) release.PilotEntries {
	m := s.Mapping
	q := fmt.Sprintf(`SELECT %s FROM %s WHERE "public_id" = $1`, toSelectClause(m), m.TableRef())
	c, err := s.ConnectionManager.Connection(ctx)
	if err != nil {
		return iterators.NewError(err)
	}

	rows, err := c.QueryContext(ctx, q, externalID)
	if err != nil {
		return iterators.NewError(err)
	}
	return iterators.NewSQLRows(rows, m)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *Postgres) ReleaseRollout(ctx context.Context) release.RolloutStorage {
	return p.storage.ReleaseRollout.Do(func() interface{} {
		return ReleaseRolloutPgStorage{
			Storage: p.mkPostgresqlStorage(release.Rollout{}, postgresql.Mapper{
				Table:   "release_rollouts",
				ID:      "id",
				NewIDFn: newIDFn,
				Columns: []string{`id`, `flag_id`, `env_id`, `plan`},
				ToArgsFn: func(ptr interface{}) ([]interface{}, error) {
					e := ptr.(*release.Rollout)
					return []interface{}{
						e.ID,
						e.FlagID,
						e.EnvironmentID,
						releaseRolloutPlanValue{RolloutPlan: e.Plan},
					}, nil
				},
				MapFn: func(s iterators.SQLRowScanner, ptr interface{}) error {
					var rollout release.Rollout
					var rolloutPlanValue releaseRolloutPlanValue
					if err := s.Scan(
						&rollout.ID,
						&rollout.FlagID,
						&rollout.EnvironmentID,
						&rolloutPlanValue,
					); err != nil {
						return err
					}

					rollout.Plan = rolloutPlanValue.RolloutPlan
					return reflects.Link(rollout, ptr)
				},
			}),
		}
	}).(ReleaseRolloutPgStorage)
}

type ReleaseRolloutPgStorage struct {
	*postgresql.Storage
}

type releaseRolloutPlanValue struct {
	release.RolloutPlan
}

func (rp releaseRolloutPlanValue) Value() (driver.Value, error) {
	return json.Marshal(release.RolloutPlanView{Plan: rp.RolloutPlan})
}

func (rp *releaseRolloutPlanValue) Scan(iSRC interface{}) error {
	src, ok := iSRC.([]byte)
	if !ok {
		const err frameless.Error = "Type assertion .([]byte) failed."
		return err
	}

	var rpv release.RolloutPlanView
	if err := json.Unmarshal(src, &rpv); err != nil {
		return err
	}

	rp.RolloutPlan = rpv.Plan
	return nil
}

func (s ReleaseRolloutPgStorage) FindByFlagEnvironment(ctx context.Context, flag release.Flag, env release.Environment, rollout *release.Rollout) (bool, error) {
	m := s.Mapping
	tmpl := `SELECT %s FROM %s WHERE flag_id = $1 AND env_id = $2`
	query := fmt.Sprintf(tmpl, toSelectClause(m), m.TableRef())

	c, err := s.ConnectionManager.Connection(ctx)
	if err != nil {
		return false, err
	}

	row := c.QueryRowContext(ctx, query, flag.ID, env.ID)
	err = m.Map(row, rollout)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *Postgres) ReleaseEnvironment(ctx context.Context) release.EnvironmentStorage {
	return p.storage.ReleaseEnvironment.Do(func() interface{} {
		return ReleaseEnvironmentPgStorage{
			Storage: p.mkPostgresqlStorage(release.Environment{}, postgresql.Mapper{
				Table:   "release_environments",
				ID:      "id",
				NewIDFn: newIDFn,
				Columns: []string{`id`, `name`},
				ToArgsFn: func(ptr interface{}) ([]interface{}, error) {
					e := ptr.(*release.Environment)
					return []interface{}{e.ID, e.Name}, nil
				},
				MapFn: func(s iterators.SQLRowScanner, ptr interface{}) error {
					e := ptr.(*release.Environment)
					return s.Scan(&e.ID, &e.Name)
				},
			}),
		}
	}).(ReleaseEnvironmentPgStorage)
}

type ReleaseEnvironmentPgStorage struct {
	*postgresql.Storage
}

func (s ReleaseEnvironmentPgStorage) FindByAlias(ctx context.Context, idOrName string, env *release.Environment) (bool, error) {
	var (
		format string
		query  string
		m      = s.Mapping
	)
	if isUUIDValid(idOrName) {
		format = `SELECT %s FROM %s WHERE id = $1`
	} else {
		format = `SELECT %s FROM %s WHERE name = $1`
	}
	query = fmt.Sprintf(format, toSelectClause(m), m.TableRef())

	c, err := s.ConnectionManager.Connection(ctx)
	if err != nil {
		return false, err
	}

	err = m.Map(c.QueryRowContext(ctx, query, idOrName), env)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (p *Postgres) SecurityToken(ctx context.Context) security.TokenStorage {
	return p.storage.SecurityToken.Do(func() interface{} {
		return SecurityTokenPgStorage{
			Storage: p.mkPostgresqlStorage(security.Token{}, postgresql.Mapper{
				Table:   "tokens", // TODO: change it to security_tokens
				ID:      "id",
				NewIDFn: newIDFn,
				Columns: []string{`id`, `sha512`, `duration`, `issued_at`, `owner_uid`},
				ToArgsFn: func(ptr interface{}) ([]interface{}, error) {
					e := ptr.(*security.Token)
					return []interface{}{
						e.ID,
						e.SHA512,
						e.Duration,
						e.IssuedAt.UTC(),
						e.OwnerUID,
					}, nil
				},
				MapFn: func(s iterators.SQLRowScanner, ptr interface{}) error {
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
					src.IssuedAt = src.IssuedAt.UTC()
					return reflects.Link(src, ptr)
				},
			}),
		}
	}).(SecurityTokenPgStorage)
}

type SecurityTokenPgStorage struct {
	*postgresql.Storage
}

func (s SecurityTokenPgStorage) FindTokenBySHA512Hex(ctx context.Context, sha512hex string) (*security.Token, error) {
	m := s.Mapping

	c, err := s.ConnectionManager.Connection(ctx)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE "sha512" = $1`, toSelectClause(m), m.TableRef())
	row := c.QueryRowContext(ctx, query, sha512hex)

	var t security.Token

	err = m.Map(row, &t)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &t, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func toSelectClause(m postgresql.Mapping) string {
	return strings.Join(m.ColumnRefs(), `,`)
}
