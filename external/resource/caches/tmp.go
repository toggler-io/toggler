package caches

//type StorageRecord struct {
//	Key     StorageRecordKey `ext:"ID"`
//	Value   interface{}
//	IsFound bool
//	IsList  bool
//}
//
//type StorageRecordKey struct {
//	T         interface{}
//	Operation string
//	Key       string
//}
//
//func ToStorageRecordKey(T interface{}, operation, key string) StorageRecordKey {
//	return StorageRecordKey{
//		T:         T,
//		Operation: operation,
//		Key:       key,
//	}
//}
//
//func (c *Manager) Subscribe() error {
//	var err error
//	c.Once.Do(func() {
//		ctx := context.Background()
//		c.exit.context, c.exit.signaler = context.WithCancel(ctx)
//
//		var sub resources.Subscription
//		for _, T := range []interface{}{
//			deployment.Environment{},
//			release.Flag{},
//			release.Rollout{},
//			release.ManualPilot{},
//			security.Token{},
//		} {
//			sub, err = c.Storage.SubscribeToCreate(ctx, T, c.CreateSubscription(T))
//			if err != nil {
//				return
//			}
//			c.eventHooks = append(c.eventHooks, sub)
//
//			sub, err = c.Storage.SubscribeToUpdate(ctx, T, c.UpdateSubscription(T))
//			if err != nil {
//				return
//			}
//			c.eventHooks = append(c.eventHooks, sub)
//
//			sub, err = c.Storage.SubscribeToDeleteByID(ctx, T, c.DeleteByIDSubscription(T))
//			if err != nil {
//				return
//			}
//			c.eventHooks = append(c.eventHooks, sub)
//
//			sub, err = c.Storage.SubscribeToDeleteAll(ctx, T, c.DeleteAllSubscription(T))
//			if err != nil {
//				return
//			}
//			c.eventHooks = append(c.eventHooks, sub)
//		}
//	})
//	return err
//}
//
//func (c *Manager) Close() error {
//	c.exit.signaler()
//	for _, sub := range c.eventHooks {
//		if err := sub.Close(); err != nil {
//			log.Println(`WARN`, err.Error())
//		}
//	}
//	_ = c.CacheStorage.Close()
//	return c.Storage.Close()
//}
//
//func (c *Manager) SetTimeToLive(duration time.Duration) error {
//	c.TTLDuration = duration
//	return nil
//}
//
//////////////////////////////////////////// cache invalidating actions //////////////////////////////////////////
//
//func (c *Manager) FindByID(ctx context.Context, ptr interface{}, id interface{}) (bool, error) {
//	return c.cacheQueryRow(ctx, ptr,
//		ToStorageRecordKey(entToT(ptr), `FindByID`, idKey(id)),
//		func(s toggler.Storage) (interface{}, bool, error) {
//			found, err := s.FindByID(ctx, ptr, id)
//			if err != nil {
//				return nil, false, err
//			}
//			return ptr, found, nil
//		})
//}
//
//func (c *Manager) FindAll(ctx context.Context, T interface{}) iterators.Interface {
//	return c.cacheQuery(ctx, T, ToStorageRecordKey(T, `FindAll`, ``),
//		func(s toggler.Storage) iterators.Interface {
//			return s.FindAll(ctx, T)
//		})
//}
//
//func (c *Manager) FindReleaseFlagByName(ctx context.Context, name string) (*release.Flag, error) {
//	var flag release.Flag
//
//	found, err := c.cacheQueryRow(ctx, &flag, ToStorageRecordKey(release.Flag{}, `FindReleaseFlagByName`, name),
//		func(s toggler.Storage) (interface{}, bool, error) {
//			flg, err := s.FindReleaseFlagByName(ctx, name)
//			if err != nil {
//				return nil, false, err
//			}
//			return flg, flg != nil, nil
//		})
//
//	if err != nil {
//		return nil, err
//	}
//
//	if !found {
//		return nil, nil
//	}
//
//	return &flag, nil
//}
//
//func (c *Manager) FindReleaseFlagsByName(ctx context.Context, names ...string) release.FlagEntries {
//	sort.Strings(names)
//	key := strings.Join(names, `,`)
//	T := release.Flag{}
//
//	return c.cacheQuery(ctx, T, ToStorageRecordKey(T, `FindReleaseFlagsByName`, key),
//		func(s toggler.Storage) iterators.Interface {
//			return s.FindReleaseFlagsByName(ctx, names...)
//		})
//}
//
//// FIXME: TODO: check if this call can be removed or manually supplied in a different manner
//// 	Caching by pilot external id potentially spawn huge number of cache record
//func (c *Manager) FindReleaseManualPilotByExternalID(ctx context.Context, flagID, envID interface{}, pilotExtID string) (*release.ManualPilot, error) {
//	var pilot release.ManualPilot
//
//	T := release.ManualPilot{}
//	key := ToStorageRecordKey(T, `FindReleaseManualPilotByExternalID`, fmt.Sprintf(`flag:%s;env:%s;pilot:%s`, idKey(flagID), idKey(envID), pilotExtID))
//	_, err := c.cacheQueryRow(ctx, &pilot, key,
//		func(s toggler.Storage) (interface{}, bool, error) {
//			pilot, err := s.FindReleaseManualPilotByExternalID(ctx, flagID, envID, pilotExtID)
//			if err != nil {
//				return nil, false, err
//			}
//			return pilot, pilot != nil, nil
//		})
//
//	return &pilot, err
//}
//
//func (c *Manager) FindReleasePilotsByReleaseFlag(ctx context.Context, flag release.Flag) release.PilotEntries {
//	T := release.ManualPilot{}
//	return c.cacheQuery(ctx, T,
//		ToStorageRecordKey(T, `FindReleasePilotsByReleaseFlag`, idKey(flag.ID)),
//		func(s toggler.Storage) iterators.Interface {
//			return s.FindReleasePilotsByReleaseFlag(ctx, flag)
//		})
//}
//
//// FIXME: TODO: check if this call can be removed actually.
//// 	This potentially will leak a huge number of empty cache hit
//func (c *Manager) FindReleasePilotsByExternalID(ctx context.Context, pilotExtID string) release.PilotEntries {
//	T := release.ManualPilot{}
//	return c.cacheQuery(ctx, T, ToStorageRecordKey(T, `FindReleasePilotsByExternalID`, pilotExtID),
//		func(s toggler.Storage) iterators.Interface {
//			return s.FindReleasePilotsByExternalID(ctx, pilotExtID)
//		})
//}
//
//func (c *Manager) FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx context.Context, flag release.Flag, environment deployment.Environment, rollout *release.Rollout) (bool, error) {
//	key := ToStorageRecordKey(release.Rollout{}, `FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment`,
//		fmt.Sprintf(`flag:%s;env:%s`, idKey(flag.ID), idKey(environment.ID)))
//
//	found, err := c.cacheQueryRow(ctx, rollout, key,
//		func(s toggler.Storage) (interface{}, bool, error) {
//			found, err := s.FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx, flag, environment, rollout)
//			if err != nil {
//				return nil, false, err
//			}
//
//			return rollout, found, nil
//		})
//	return found, err
//}
//
//func (c *Manager) FindDeploymentEnvironmentByAlias(ctx context.Context, idOrName string, env *deployment.Environment) (bool, error) {
//	key := ToStorageRecordKey(deployment.Environment{}, `FindDeploymentEnvironmentByAlias`, idOrName)
//	found, err := c.cacheQueryRow(ctx, env, key,
//		func(s toggler.Storage) (interface{}, bool, error) {
//			found, err := s.FindDeploymentEnvironmentByAlias(ctx, idOrName, env)
//			if err != nil {
//				return nil, false, err
//			}
//			return env, found, nil
//		})
//	return found, err
//}
//
//type skipCacheCtxKey struct{}
//
//func (c *Manager) BeginTx(ctx context.Context) (context.Context, error) {
//	ctx, err := c.Storage.BeginTx(ctx)
//	if err != nil {
//		return ctx, err
//	}
//
//	return context.WithValue(ctx, skipCacheCtxKey{}, struct{}{}), nil
//}
//
//// ----------------------------------------------------- caching ---------------------------------------------------- //
//
//func (c *Manager) invalidateByID(ctx context.Context, T, id interface{}) error {
//	allSR := c.CacheStorage.FindAll(context.Background(), StorageRecord{})
//	filtered := iterators.Filter(allSR, func(r StorageRecord) bool {
//		if r.Key.T != T {
//			return false
//		}
//
//		var isIDFound bool
//
//		if r.IsList {
//			_ = iterators.ForEach(iterators.NewSlice(r.Value), func(i interface{}) error {
//				if rid, ok := resources.LookupID(i); ok && rid == id {
//					isIDFound = true
//					return iterators.Break
//				}
//
//				return nil
//			})
//		} else {
//			if rid, ok := resources.LookupID(r.Value); ok && rid == id {
//				isIDFound = true
//			}
//		}
//
//		return isIDFound
//	})
//
//	var srs []StorageRecord
//	if err := iterators.Collect(filtered, &srs); err != nil {
//		return err
//	}
//
//	for _, sr := range srs {
//		if err := c.CacheStorage.DeleteByID(ctx, T, sr.Key); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (c *Manager) invalidateAll(T interface{}) error {
//	ctx := context.Background()
//
//	allSR := c.CacheStorage.FindAll(context.Background(), StorageRecord{})
//	filtered := iterators.Filter(allSR, func(r StorageRecord) bool {
//		return r.Key.T == T
//	})
//
//	var srs []StorageRecord
//	if err := iterators.Collect(filtered, &srs); err != nil {
//		return err
//	}
//
//	for _, sr := range srs {
//		if err := c.CacheStorage.DeleteByID(ctx, T, sr.Key); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (c *Manager) prolongTTL(ctx context.Context, ents ...interface{}) {
//	for _, ent := range ents {
//		T := entToT(ent)
//		id, _ := resources.LookupID(ent)
//		_ = c.ttl(ctx, T, id)
//	}
//}
//
//func (c *Manager) ttl(ctx context.Context, T interface{}, id interface{}) error {
//	if c.TTLDuration == 0 {
//		return nil
//	}
//
//	return c.CacheStorage.TTL(ctx, T, id, c.TTLDuration)
//}
//
//func (c *Manager) cacheQueryRow(ctx context.Context, ptr interface{}, key StorageRecordKey, query func(s toggler.Storage) (interface{}, bool, error)) (bool, error) {
//	var fetchAndLink = func() (interface{}, bool, error) {
//		value, found, err := query(c.Storage)
//		if err != nil {
//			return nil, false, err
//		}
//
//		if found {
//			if err := reflects.Link(reflects.BaseValueOf(value).Interface(), ptr); err != nil {
//				return nil, false, err
//			}
//		}
//
//		return value, found, err
//	}
//
//	if ctx.Value(skipCacheCtxKey{}) != nil {
//		_, found, err := fetchAndLink()
//		return found, err
//	}
//
//	var r StorageRecord
//
//	if found, err := c.CacheStorage.FindByID(ctx, &r, key); err != nil {
//		_, found, err := fetchAndLink()
//
//		return found, err
//	} else if found {
//		if !r.IsFound {
//			return false, nil
//		}
//
//		return true, reflects.Link(reflects.BaseValueOf(r.Value).Interface(), ptr)
//	}
//
//	value, found, err := fetchAndLink()
//	if err != nil {
//		return false, err
//	}
//
//	if err := c.CacheStorage.Create(ctx, &StorageRecord{
//		Key:     key,
//		Value:   value,
//		IsFound: found,
//	}); err == nil {
//		_ = c.ttl(ctx, StorageRecord{}, key)
//	} else {
//		log.Println(`WARN`, `caching query failed: `, err.Error())
//	}
//
//	return found, nil
//}
//
//func (c *Manager) cacheQuery(ctx context.Context, T resources.T, key StorageRecordKey, query func(s toggler.Storage) iterators.Interface) iterators.Interface {
//	if ctx.Value(skipCacheCtxKey{}) != nil {
//		return query(c.Storage)
//	}
//
//	var r StorageRecord
//
//	if found, err := c.CacheStorage.FindByID(ctx, &r, key); err != nil {
//		return query(c.Storage)
//
//	} else if found {
//		if !r.IsFound {
//			return iterators.NewEmpty()
//		}
//
//		return iterators.NewSlice(r.Value)
//	}
//
//	rSliceType := reflect.SliceOf(reflect.TypeOf(T))
//	ptrToSlice := reflect.New(rSliceType)
//	ptrToSlice.Elem().Set(reflect.MakeSlice(rSliceType, 0, 0))
//
//	if err := iterators.Collect(query(c.Storage), ptrToSlice.Interface()); err != nil {
//		return iterators.NewError(err)
//	}
//
//	if err := c.CacheStorage.Create(ctx, &StorageRecord{
//		Key:     key,
//		Value:   ptrToSlice.Elem().Interface(),
//		IsFound: true,
//		IsList:  true,
//	}); err == nil {
//		_ = c.ttl(ctx, StorageRecord{}, key)
//	} else {
//		log.Println(`WARN`, `caching query failed: `, err.Error())
//	}
//
//	return iterators.NewSlice(reflect.MakeSlice(rSliceType, 0, 0).Interface())
//}
//
//type MemorySubscriber struct {
//	HandleFunc func(ctx context.Context, T interface{}) error
//	ErrorFunc  func(ctx context.Context, err error) error
//}
//
//func (m MemorySubscriber) Handle(ctx context.Context, T interface{}) error {
//	if m.HandleFunc == nil {
//		return nil
//	}
//
//	return m.HandleFunc(ctx, T)
//}
//
//func (m MemorySubscriber) Error(ctx context.Context, err error) error {
//	if m.ErrorFunc == nil {
//		return nil
//	}
//
//	return m.ErrorFunc(ctx, err)
//}
//
//func (c *Manager) CreateSubscription(T interface{}) resources.Subscriber {
//	return MemorySubscriber{
//		HandleFunc: func(ctx context.Context, e interface{}) error {
//			id, ok := resources.LookupID(e)
//			if !ok {
//				return c.invalidateAll(T)
//			}
//
//			if err := c.invalidateByID(ctx, T, id); err != nil {
//				return err
//			}
//
//			if err := c.CacheStorage.DeleteStorageRecordsByTypeAndOperation(ctx, T, `FindAll`); err != nil {
//				return err
//			}
//
//			return nil
//		},
//		ErrorFunc: func(ctx context.Context, err error) error {
//			return c.invalidateAll(T)
//		},
//	}
//}
//
//func (c *Manager) UpdateSubscription(T interface{}) resources.Subscriber {
//	return MemorySubscriber{
//		HandleFunc: func(ctx context.Context, e interface{}) error {
//			id, _ := resources.LookupID(e)
//			return c.invalidateByID(ctx, T, id)
//		},
//		ErrorFunc: func(ctx context.Context, err error) error {
//			return c.invalidateAll(T)
//		},
//	}
//}
//
//func (c *Manager) DeleteAllSubscription(T interface{}) resources.Subscriber {
//	return MemorySubscriber{
//		HandleFunc: func(ctx context.Context, e interface{}) error {
//			return c.invalidateAll(T)
//		},
//		ErrorFunc: func(ctx context.Context, err error) error {
//			return c.invalidateAll(T)
//		},
//	}
//}
//
//func (c *Manager) DeleteByIDSubscription(T interface{}) resources.Subscriber {
//	return MemorySubscriber{
//		HandleFunc: func(ctx context.Context, e interface{}) error {
//			id, _ := resources.LookupID(e)
//			return c.invalidateByID(ctx, T, id)
//		},
//		ErrorFunc: func(ctx context.Context, err error) error {
//			return c.invalidateAll(T)
//		},
//	}
//}
//
//func entToT(ptr interface{}) (T interface{}) {
//	return reflects.BaseValueOf(reflect.New(reflects.BaseTypeOf(ptr))).Interface()
//}
//
//func tKey(T interface{}) string {
//	return reflects.FullyQualifiedName(T)
//}
//
//func idKey(id interface{}) string {
//	switch id := id.(type) {
//	case string:
//		return id
//	default:
//		return fmt.Sprintf(`%#v`, id)
//	}
//}
