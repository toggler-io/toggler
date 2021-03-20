package caches

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"sync"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources"
	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/domains/toggler"
)

func NewManager(togglerStorage toggler.Storage, cacheStorage Storage) (*Manager, error) {
	m := &Manager{
		Storage:                   togglerStorage,
		CacheStorage:              cacheStorage,
		EntityTypeNameMappingFunc: defaultEntityTypeNameMapping,
		CachedEntityTypes:         cachedEntityTypes,
	}
	return m, m.Subscribe()
}

type Manager struct {
	toggler.Storage
	CacheStorage Storage
	EntityTypeNameMappingFunc
	CachedEntityTypes []interface{}

	shutdown struct {
		sync.Once
		subscription func() error
	}
}

type Storage /* [T] */ interface {
	io.Closer
	resources.Creator
	resources.Finder
	resources.Updater
	resources.Deleter

	UpsertMany(ctx context.Context, ptrs ...interface{}) error
	FindByIDs(ctx context.Context, T resources.T, ids ...interface{}) iterators.Interface /* [T] */
}

type inTxCtxKey struct{}

func (m *Manager) isInTx(ctx context.Context) bool {
	return ctx.Value(inTxCtxKey{}) != nil
}

func (m *Manager) BeginTx(ctx context.Context) (context.Context, error) {
	return m.Storage.BeginTx(context.WithValue(ctx, inTxCtxKey{}, struct{}{}))
}

func (m *Manager) Close() error {
	var subErr, cacheStorageErr, storageErr error
	if s := m.shutdown.subscription; s != nil {
		subErr = s()
	}
	cacheStorageErr = m.CacheStorage.Close()
	storageErr = m.Storage.Close()
	if storageErr != nil {
		return storageErr
	}
	if subErr != nil {
		return subErr
	}
	if cacheStorageErr != nil {
		return cacheStorageErr
	}
	return nil
}

//----------------------------------------------- QueryOne & QueryMany -----------------------------------------------//

type CachedQuery struct {
	// ID is the encoded query request identifier
	ID string `ext:"ID"`
	// EntityType is the enum value that can be used to know the entity type
	EntityType string
	// HitIDs are the id list
	HitIDs []interface{}
}

func (m *Manager) cacheQuery(
	ctx context.Context,
	queryID string,
	T interface{},
	query func() iterators.Interface,
) iterators.Interface {
	if m.isInTx(ctx) {
		return query()
	}

	name := m.EntityTypeNameMappingFunc(T)
	queryID = fmt.Sprintf(`0:%s/%s`, name, queryID)
	var q CachedQuery
	found, err := m.CacheStorage.FindByID(ctx, &q, queryID)
	if err != nil {
		return iterators.NewError(err)
	}
	if found {
		return m.CacheStorage.FindByIDs(ctx, T, q.HitIDs...)
	}

	// this naive MVP approach might take a big burden on the memory.
	// If this becomes the case, it should be possible to change this into a streaming approach
	// where iterator being iterated element by element,
	// and records being created during then in the CacheStorage
	var vs, ids []interface{}
	if err := iterators.Collect(query(), &vs); err != nil {
		return iterators.NewError(err)
	}
	for _, v := range vs {
		id, _ := resources.LookupID(v)
		ids = append(ids, id)
	}

	if err := m.CacheStorage.UpsertMany(ctx, vs...); err != nil {
		return iterators.NewError(err)
	}

	if err := m.CacheStorage.Create(ctx, &CachedQuery{
		ID:         queryID,
		EntityType: m.EntityTypeNameMappingFunc(T),
		HitIDs:     ids,
	}); err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSlice(vs)
}

func (m *Manager) cacheQueryOne(
	ctx context.Context,
	queryID string,
	T interface{},
	ptr interface{},
	query func() (v interface{}, err error),
) (_found bool, _err error) {
	return iterators.First(m.cacheQuery(ctx, queryID, T, func() iterators.Interface {
		v, err := query()
		if err != nil {
			return iterators.NewError(err)
		}
		if v == nil {
			return iterators.NewEmpty()
		}
		if reflect.ValueOf(v).IsZero() {
			// interface{} Has (*ptrType)(nil)
			// since interface{} points to a pointer type that points to a nil
			// the interface{} content is not nil but a pointer type.
			return iterators.NewEmpty()
		}
		return iterators.NewSlice([]interface{}{v})
	}), ptr)
}

func (m *Manager) FindByID(ctx context.Context, ptr, id interface{}) (bool, error) {
	if m.isInTx(ctx) {
		return m.Storage.FindByID(ctx, ptr, id)
	}

	found, err := m.CacheStorage.FindByID(ctx, ptr, id)
	if err != nil {
		return false, err
	}
	if found {
		return found, nil
	}
	found, err = m.Storage.FindByID(ctx, ptr, id)
	if err != nil {
		return false, err
	}
	if found {
		if err := m.CacheStorage.Create(ctx, ptr); err != nil {
			return false, err
		}
	}
	return found, nil
}

func (m *Manager) FindAll(ctx context.Context, T interface{}) iterators.Interface {
	entityType := m.EntityTypeNameMappingFunc(T)
	queryID := fmt.Sprintf(`FindAll{T:%s}`, entityType)
	return m.cacheQuery(ctx, queryID, T, func() iterators.Interface {
		return m.Storage.FindAll(ctx, T)
	})
}

func (m *Manager) FindReleaseFlagByName(ctx context.Context, name string) (*release.Flag, error) {
	queryID := fmt.Sprintf(`FindReleaseFlagByName{name:%s}`, name)
	f := &release.Flag{}
	found, err := m.cacheQueryOne(ctx, queryID, release.Flag{}, &f, func() (v interface{}, err error) {
		return m.Storage.FindReleaseFlagByName(ctx, name)
	})
	if err != nil {
		return nil, err
	}
	if !found {
		f = nil
	}
	return f, nil
}

func (m *Manager) FindReleaseFlagsByName(ctx context.Context, names ...string) release.FlagEntries {
	sort.Strings(names)
	fingerprintBS, err := json.Marshal(names)
	if err != nil {
		return iterators.NewError(err)
	}
	queryID := fmt.Sprintf(`FindReleaseFlagsByName{names:%s}`, fingerprintBS)

	return m.cacheQuery(ctx, queryID, release.Flag{}, func() iterators.Interface {
		return m.Storage.FindReleaseFlagsByName(ctx, names...)
	})
}

func (m *Manager) FindReleaseManualPilotByExternalID(ctx context.Context, flagID, envID interface{}, pilotExtID string) (*release.ManualPilot, error) {
	queryID := fmt.Sprintf(`FindReleaseManualPilotByExternalID{flagID:%s,envID:%s,pilotExtID:%s}`, flagID, envID, pilotExtID)
	p := &release.ManualPilot{}
	found, err := m.cacheQueryOne(ctx, queryID, release.ManualPilot{}, &p, func() (v interface{}, err error) {
		// TODO: if nil returned to the cache, should it be used?
		return m.Storage.FindReleaseManualPilotByExternalID(ctx, flagID, envID, pilotExtID)
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return p, nil
}

func (m *Manager) FindReleasePilotsByReleaseFlag(ctx context.Context, flag release.Flag) release.PilotEntries {
	queryID := fmt.Sprintf(`FindReleasePilotsByReleaseFlag{flagID:%s}`, flag.ID)
	return m.cacheQuery(ctx, queryID, release.ManualPilot{}, func() iterators.Interface {
		return m.Storage.FindReleasePilotsByReleaseFlag(ctx, flag)
	})
}

func (m *Manager) FindReleasePilotsByExternalID(ctx context.Context, externalID string) release.PilotEntries {
	queryID := fmt.Sprintf(`FindReleasePilotsByExternalID{pilotExtID:%s}`, externalID)
	return m.cacheQuery(ctx, queryID, release.ManualPilot{}, func() iterators.Interface {
		return m.Storage.FindReleasePilotsByExternalID(ctx, externalID)
	})
}

func (m *Manager) FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx context.Context, flag release.Flag, environment deployment.Environment, rollout *release.Rollout) (bool, error) {
	queryID := fmt.Sprintf(`FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment{flagID:%s,envID:%s}`, flag.ID, environment.ID)
	return m.cacheQueryOne(ctx, queryID, release.Rollout{}, rollout, func() (v interface{}, err error) {
		var r release.Rollout
		found, err := m.Storage.FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx, flag, environment, &r)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, err
		}
		return r, err
	})
}

//func (m *Manager) FindTokenBySHA512Hex(ctx context.Context, sha512hex string) (*security.Token, error) {
//	panic("implement me")
//}
//
//func (m *Manager) FindDeploymentEnvironmentByAlias(ctx context.Context, idOrName string, env *deployment.Environment) (bool, error) {
//	panic("implement me")
//}

//--------------------------------------------------- subscription ---------------------------------------------------//

var cachedEntityTypes = []interface{}{
	release.Flag{},
	release.Rollout{},
	release.ManualPilot{},
	deployment.Environment{},
	security.Token{},
}

func (m *Manager) Subscribe() error {
	var err error
	m.shutdown.Once.Do(func() {
		ctx := context.Background()

		var subscriptions []resources.Subscription
		for _, T := range m.CachedEntityTypes {
			var subscription resources.Subscription

			subscription, err = m.Storage.SubscribeToDeleteByID(ctx, T, m.getSubscriberDeleteByID(T))
			if err != nil {
				return
			}
			subscriptions = append(subscriptions, subscription)

			subscription, err = m.Storage.SubscribeToDeleteAll(ctx, T, m.getSubscriberDeleteAll(T))
			if err != nil {
				return
			}
			subscriptions = append(subscriptions, subscription)

			subscription, err = m.Storage.SubscribeToUpdate(ctx, T, m.getSubscriberUpdate(T))
			if err != nil {
				return
			}
			subscriptions = append(subscriptions, subscription)

			subscription, err = m.Storage.SubscribeToCreate(ctx, T, m.getSubscriberCreate(T))
			if err != nil {
				return
			}
			subscriptions = append(subscriptions, subscription)
		}

		m.shutdown.subscription = func() error {
			m.shutdown.Once = sync.Once{}
			var rErr error
			for _, sub := range subscriptions {
				if err := sub.Close(); err != nil {
					rErr = err
				}
			}
			return rErr
		}
	})
	return err
}

type managerSubscriber struct {
	HandleFunc func(ctx context.Context, ent interface{}) error
	ErrorFunc  func(ctx context.Context, err error) error
}

func (m managerSubscriber) Handle(ctx context.Context, ent interface{}) error {
	if m.HandleFunc != nil {
		return m.HandleFunc(ctx, ent)
	}

	return nil
}

func (m managerSubscriber) Error(ctx context.Context, err error) error {
	if m.ErrorFunc != nil {
		return m.ErrorFunc(ctx, err)
	}

	return nil
}

func (m *Manager) entCachedQueries(ctx context.Context, T interface{}) *iterators.FilterIterator {
	etn := m.EntityTypeNameMappingFunc(T)
	all := m.CacheStorage.FindAll(ctx, CachedQuery{})
	return iterators.Filter(all, func(r CachedQuery) bool {
		return r.EntityType == etn
	})
}

func (m *Manager) deleteCachedQueryFor(ctx context.Context, T interface{}) error {
	var cqs []CachedQuery
	if err := iterators.Collect(m.entCachedQueries(ctx, T), &cqs); err != nil {
		return err
	}

	for _, cq := range cqs {
		if err := m.CacheStorage.DeleteByID(ctx, CachedQuery{}, cq.ID); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) deleteCachedEntity(ctx context.Context, T, id interface{}) error {
	ptr := reflect.New(reflect.TypeOf(T)).Interface()
	found, err := m.CacheStorage.FindByID(ctx, ptr, id)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	return m.CacheStorage.DeleteByID(ctx, T, id)
}

func (m *Manager) getSubscriberCreate(T interface{}) resources.Subscriber {
	return managerSubscriber{
		HandleFunc: func(ctx context.Context, ent interface{}) error {
			return m.deleteCachedQueryFor(ctx, T)
		},
		ErrorFunc: func(ctx context.Context, err error) error {
			return m.CacheStorage.DeleteAll(ctx, CachedQuery{})
		},
	}
}

func (m *Manager) getSubscriberUpdate(T interface{}) resources.Subscriber {
	return managerSubscriber{
		HandleFunc: func(ctx context.Context, ent interface{}) error {
			if err := m.deleteCachedQueryFor(ctx, T); err != nil {
				return err
			}
			id, _ := resources.LookupID(ent)
			return m.deleteCachedEntity(ctx, T, id)
		},
		ErrorFunc: func(ctx context.Context, err error) error {
			if err := m.CacheStorage.DeleteAll(ctx, T); err != nil {
				return err
			}

			if err := m.CacheStorage.DeleteAll(ctx, CachedQuery{}); err != nil {
				return err
			}

			return nil
		},
	}
}

func (m *Manager) getSubscriberDeleteAll(T interface{}) resources.Subscriber {
	return managerSubscriber{
		HandleFunc: func(ctx context.Context, ent interface{}) error {
			if err := m.deleteCachedQueryFor(ctx, T); err != nil {
				return err
			}

			return m.CacheStorage.DeleteAll(ctx, T)
		},
		ErrorFunc: func(ctx context.Context, err error) error {
			return m.CacheStorage.DeleteAll(ctx, T)
		},
	}
}

func (m *Manager) getSubscriberDeleteByID(T interface{}) resources.Subscriber {
	return managerSubscriber{
		HandleFunc: func(ctx context.Context, ent interface{}) error {
			if err := m.deleteCachedQueryFor(ctx, T); err != nil {
				return err
			}
			id, _ := resources.LookupID(ent)
			return m.deleteCachedEntity(ctx, T, id)
		},
		ErrorFunc: func(ctx context.Context, err error) error {
			return m.CacheStorage.DeleteAll(ctx, T)
		},
	}
}
