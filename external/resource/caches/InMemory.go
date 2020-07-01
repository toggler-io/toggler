package caches

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/usecases"
)

func NewInMemory(s usecases.Storage) *InMemory {
	c := &InMemory{Storage: s, ttl: 5 * time.Minute}
	c.Start()
	return c
}

//TODO: possible improvement to protect with mux around actions instead of set and lookup, so on concurrent access there will be only 1 find
//TODO: implement cache invalidation with delete/update event streams in the next iterations
type InMemory struct {
	Storage usecases.Storage
	cache   map[string]map[string]*inMemoryCachedItem
	ttl     time.Duration

	lock   sync.Mutex
	init   sync.Once
	cancel func()
}

func (c *InMemory) SetTimeToLiveForValuesToCache(duration time.Duration) error {
	c.ttl = duration
	return nil
}

type inMemoryCachedItem struct {
	value   interface{}
	updater updater

	createdAt   time.Time
	lastAccess  time.Time
	lastUpdated time.Time
}

func (c *InMemory) Start() {
	c.init.Do(func() {
		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())
		c.cancel = func() {
			cancel()
			wg.Wait()
		}

		wg.Add(+1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case <-time.Tick(time.Second):
				c.gcWRK()
			}
		}()

		wg.Add(+1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case <-time.Tick(time.Minute):
				c.updateCacheItems()
			}
		}()
	})
}

func (c *InMemory) updateCacheItems() {
	for _, nsv := range c.cache {
		for _, item := range nsv {
			if err := item.updater(setterFuncWrapper(func(value interface{}) {
				item.value = value
				item.lastUpdated = time.Now()
			})); err != nil {
			}
		}
	}
}

func (c *InMemory) gcWRK() {
	now := time.Now()
	for _, nsv := range c.cache {
		for key, item := range nsv {
			// for now, using TTL on the created at time is enough,
			// since we don't have push event about item deletion yet.
			if now.Add(c.ttl * -1).After(item.lastAccess) {
				c.lock.Lock()
				delete(nsv, key)
				c.lock.Unlock()
			}
		}
	}
}

func (c *InMemory) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return c.Storage.Close()
}

type updater func(setter) error

type setter interface {
	set(value interface{})
}

type setterFuncWrapper func(value interface{})

func (fn setterFuncWrapper) set(value interface{}) {
	fn(value)
}

func (c *InMemory) get(namespace, key string, upd updater) (interface{}, error) {
	v, ok := c.lookup(namespace, key)
	if !ok {
		if err := upd(setterFuncWrapper(func(newValue interface{}) {
			item, ok := c.namespace(namespace)[key]
			if !ok {
				item = &inMemoryCachedItem{value: newValue, updater: upd}
				c.namespace(namespace)[key] = item
				item.createdAt = time.Now()
			}
			item.lastAccess = time.Now()
		})); err != nil {
			return nil, err
		}
	}
	v, _ = c.lookup(namespace, key)
	return v, nil
}

func (c *InMemory) namespaceKey(T interface{}) string {
	switch T.(type) {
	case release.Flag, *release.Flag:
		return `release.Flag`
	case release.ManualPilot, *release.ManualPilot:
		return `release.ManualPilot`
	case security.Token:
		return `security.Token`
	default:
		return reflects.SymbolicName(T)
	}
}

func (c *InMemory) namespace(namespaceKey string) map[string]*inMemoryCachedItem {
	if c.cache == nil {
		c.cache = make(map[string]map[string]*inMemoryCachedItem)
	}
	if c.cache[namespaceKey] == nil {
		c.cache[namespaceKey] = make(map[string]*inMemoryCachedItem)
	}
	return c.cache[namespaceKey]
}

func (c *InMemory) dropCache() {
	c.cache = nil
}

func (c *InMemory) lookup(namespace, key string) (interface{}, bool) {
	var (
		value interface{}
		found bool
	)
	if item, ok := c.namespace(namespace)[key]; ok {
		value = item.value
		item.lastAccess = time.Now()
		found = true
	}
	return value, found
}

func (c *InMemory) withLock() func() {
	c.lock.Lock()
	return c.lock.Unlock
}

////////////////////////////////////////// cached actions //////////////////////////////////////////

func (c *InMemory) Create(ctx context.Context, value interface{}) error {
	defer c.withLock()()
	c.dropCache()
	return c.Storage.Create(ctx, value)
}

func (c *InMemory) FindByID(ctx context.Context, ptr interface{}, id string) (_found bool, _err error) {
	defer c.withLock()()

	if shouldSkipCache(ctx) {
		return c.Storage.FindByID(ctx, ptr, id)
	}

	type ValueOfFindByID struct {
		value interface{}
		found bool
	}

	v, err := c.get(c.namespaceKey(ptr), id, func(s setter) error {
		found, err := c.Storage.FindByID(ctx, ptr, id)
		if err != nil {
			return err
		}

		s.set(&ValueOfFindByID{value: reflects.BaseValueOf(ptr).Interface(), found: found})
		return nil
	})

	if err != nil {
		return false, err
	}

	fbii := v.(*ValueOfFindByID)
	if !fbii.found {
		return false, nil
	}

	return true, reflects.Link(fbii.value, ptr)
}

func (c *InMemory) FindAll(ctx context.Context, T interface{}) frameless.Iterator {
	defer c.withLock()()

	if shouldSkipCache(ctx) {
		return c.Storage.FindAll(ctx, T)
	}

	const namespace = `FindAll`
	v, err := c.get(namespace, c.namespaceKey(T), func(s setter) error {
		iter := c.Storage.FindAll(ctx, T)
		var results []interface{}
		if err := iterators.Collect(iter, &results); err != nil {
			return err
		}
		s.set(results)
		return nil
	})

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSlice(v)
}

func (c *InMemory) Update(ctx context.Context, ptr interface{}) error {
	defer c.withLock()()
	c.dropCache()
	return c.Storage.Update(ctx, ptr)
}

func (c *InMemory) DeleteByID(ctx context.Context, T interface{}, id string) error {
	defer c.withLock()()
	c.dropCache()
	return c.Storage.DeleteByID(ctx, T, id)
}

func (c *InMemory) DeleteAll(ctx context.Context, T interface{}) error {
	defer c.withLock()()
	c.dropCache()
	return c.Storage.DeleteAll(ctx, T)
}

func (c *InMemory) FindReleaseFlagByName(ctx context.Context, name string) (*release.Flag, error) {
	defer c.withLock()()

	if shouldSkipCache(ctx) {
		return c.Storage.FindReleaseFlagByName(ctx, name)
	}

	const namespace = `FindReleaseFlagByName`

	v, err := c.get(namespace, name, func(s setter) error {
		ff, err := c.Storage.FindReleaseFlagByName(ctx, name)
		if err != nil {
			return err
		}
		s.set(ff)
		return nil
	})
	return v.(*release.Flag), err
}

func (c *InMemory) FindReleaseFlagsByName(ctx context.Context, names ...string) release.FlagEntries {
	defer c.withLock()()

	if shouldSkipCache(ctx) {
		return c.Storage.FindReleaseFlagsByName(ctx, names...)
	}

	const namespace = `FindReleaseFlagsByName`

	sort.Strings(names)
	key := strings.Join(names, `.`)

	v, err := c.get(namespace, key, func(s setter) error {
		var flags []interface{}
		if err := iterators.Collect(c.Storage.FindReleaseFlagsByName(ctx, names...), &flags); err != nil {
			return err
		}
		s.set(flags)
		return nil
	})

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSlice(v)
}

func (c *InMemory) FindReleaseManualPilotByExternalID(ctx context.Context, flagID, envID, pilotExtID string) (*release.ManualPilot, error) {
	defer c.withLock()()

	if shouldSkipCache(ctx) {
		return c.Storage.FindReleaseManualPilotByExternalID(ctx, flagID, envID, pilotExtID)
	}

	const namespace = `FindReleaseManualPilotByExternalID`
	var key = fmt.Sprintf(`%s|%s|%s`, flagID, envID, pilotExtID)
	v, err := c.get(namespace, key, func(s setter) error {
		p, err := c.Storage.FindReleaseManualPilotByExternalID(ctx, flagID, envID, pilotExtID)
		if err != nil {
			return err
		}
		s.set(p)
		return nil
	})
	return v.(*release.ManualPilot), err
}

func (c *InMemory) FindReleasePilotsByReleaseFlag(ctx context.Context, flag release.Flag) release.PilotEntries {
	defer c.withLock()()

	if shouldSkipCache(ctx) {
		return c.Storage.FindReleasePilotsByReleaseFlag(ctx, flag)
	}

	const namespace = `FindReleasePilotsByReleaseFlag`

	if flag.ID == `` {
		return iterators.NewEmpty()
	}

	if id, _ := resources.LookupID(flag); id == `` {
		return iterators.NewEmpty()
	}

	v, err := c.get(namespace, flag.ID, func(s setter) error {
		var pilots []interface{}
		if err := iterators.Collect(c.Storage.FindReleasePilotsByReleaseFlag(ctx, flag), &pilots); err != nil {
			return err
		}
		s.set(pilots)
		return nil
	})

	if err != nil {
		return iterators.NewError(err)
	}

	return iterators.NewSlice(v)
}

func (c *InMemory) FindReleasePilotsByExternalID(ctx context.Context, pilotExtID string) release.PilotEntries {
	defer c.withLock()()

	if shouldSkipCache(ctx) {
		return c.Storage.FindReleasePilotsByExternalID(ctx, pilotExtID)
	}

	const namespace = `FindReleasePilotsByExternalID`

	v, err := c.get(namespace, pilotExtID, func(s setter) error {
		var pilots []interface{}
		if err := iterators.Collect(c.Storage.FindReleasePilotsByExternalID(ctx, pilotExtID), &pilots); err != nil {
			return err
		}
		s.set(pilots)
		return nil
	})

	if err != nil {
		return iterators.NewError(err)
	}
	return iterators.NewSlice(v)
}

func (c *InMemory) FindTokenBySHA512Hex(ctx context.Context, sha512hex string) (*security.Token, error) {
	defer c.withLock()()

	if shouldSkipCache(ctx) {
		return c.Storage.FindTokenBySHA512Hex(ctx, sha512hex)
	}

	const namespace = `FindTokenBySHA512Hex`
	v, err := c.get(namespace, sha512hex, func(s setter) error {
		t, err := c.Storage.FindTokenBySHA512Hex(ctx, sha512hex)
		if err != nil {
			return err
		}
		s.set(t)
		return nil
	})
	return v.(*security.Token), err
}

func (c *InMemory) FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx context.Context, flag release.Flag, environment deployment.Environment, rollout *release.Rollout) (bool, error) {
	defer c.withLock()()

	if shouldSkipCache(ctx) {
		return c.Storage.FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx, flag, environment, rollout)
	}

	type FindReleaseRolloutByReleaseFlagAndDeploymentEnvironmentValue struct {
		value release.Rollout
		found bool
	}

	key := strings.Join([]string{`flag`, flag.ID, `env`, environment.ID}, `/`)

	v, err := c.get(c.namespaceKey(FindReleaseRolloutByReleaseFlagAndDeploymentEnvironmentValue{}), key, func(s setter) error {
		var r release.Rollout
		found, err := c.Storage.FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx, flag, environment, &r)
		if err != nil {
			return err
		}

		s.set(&FindReleaseRolloutByReleaseFlagAndDeploymentEnvironmentValue{value: r, found: found})
		return nil
	})

	if err != nil {
		return false, err
	}

	result := v.(*FindReleaseRolloutByReleaseFlagAndDeploymentEnvironmentValue)
	if !result.found {
		return false, nil
	}

	return true, reflects.Link(result.value, rollout)
}

func (c *InMemory) FindDeploymentEnvironmentByAlias(ctx context.Context, idOrName string, env *deployment.Environment) (bool, error) {
	defer c.withLock()()

	if shouldSkipCache(ctx) {
		return c.Storage.FindDeploymentEnvironmentByAlias(ctx, idOrName, env)
	}

	type FindDeploymentEnvironmentByAliasValue struct {
		value deployment.Environment
		found bool
	}

	key := strings.Join([]string{`idOrName`, idOrName, `env`, env.ID}, `/`)

	v, err := c.get(c.namespaceKey(FindDeploymentEnvironmentByAliasValue{}), key, func(s setter) error {
		var d deployment.Environment
		found, err := c.Storage.FindDeploymentEnvironmentByAlias(ctx, idOrName, &d)
		if err != nil {
			return err
		}

		s.set(&FindDeploymentEnvironmentByAliasValue{value: d, found: found})
		return nil
	})

	if err != nil {
		return false, err
	}

	value := v.(*FindDeploymentEnvironmentByAliasValue)
	if !value.found {
		return false, nil
	}

	return true, reflects.Link(value.value, env)
}

func (c *InMemory) BeginTx(ctx context.Context) (context.Context, error) {
	return c.Storage.BeginTx(contextWithNoCache(ctx))
}

func (c *InMemory) CommitTx(ctx context.Context) error {
	if err := c.Storage.CommitTx(ctx); err != nil {
		return err
	}

	noCacheDone(ctx)
	c.dropCache()
	return nil
}

func (c *InMemory) RollbackTx(ctx context.Context) error {
	if err := c.Storage.RollbackTx(ctx); err != nil {
		return err
	}

	noCacheDone(ctx)
	c.dropCache()
	return nil
}
