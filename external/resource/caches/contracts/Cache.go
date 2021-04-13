//go:generate mockgen -package contracts -source ../../../../domains/toggler/Storage.go -destination MockStorage.go
package contracts

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources"
	contracts2 "github.com/adamluzsi/frameless/resources/contracts"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/domains/toggler/contracts"
	sh "github.com/toggler-io/toggler/spechelper"
)

var (
	waiter = testcase.Waiter{
		WaitDuration: time.Millisecond,
		WaitTimeout:  time.Second,
	}
	async = testcase.Retry{Strategy: &waiter}
)

type Cache struct {
	NewCache       func(testing.TB, toggler.Storage) toggler.Storage
	FixtureFactory sh.FixtureFactory
}

func (spec Cache) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec Cache) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec Cache) Spec(tb testing.TB) {
	testcase.NewSpec(tb).Describe(`Cache`, func(s *testcase.Spec) {
		spec.setup(s)

		s.Test(`supplies toggler.Storage contract`, func(t *testcase.T) {
			testcase.RunContract(t, contracts.Storage{Subject: func(tb testing.TB) toggler.Storage {
				return spec.cacheGet(t)
			}})
		})

		Ts := []interface{}{
			deployment.Environment{},
			release.Flag{},
			release.Rollout{},
			release.ManualPilot{},
		}

		s.Describe(`query results caching`, func(s *testcase.Spec) {
			for _, T := range Ts {
				spec.expectResultCachingFor(s, T)
			}
		})

		s.Describe(`cache invalidation by events that mutates an entity`, func(s *testcase.Spec) {
			for _, T := range Ts {
				spec.specCacheInvalidationByEventsThatMutatesAnEntity(s, T)
			}
		})
	})
}

func (spec Cache) setup(s *testcase.Spec) {
	sh.SetUp(s) // Cache depends on spechelper.Storage Var
	spec.cache().Let(s, nil)
	spec.storage().Let(s, nil)
}

func (spec Cache) specCacheInvalidationByEventsThatMutatesAnEntity(s *testcase.Spec, T interface{}) {
	s.Context(reflects.SymbolicName(T), func(s *testcase.Spec) {
		s.Let(`value`, func(t *testcase.T) interface{} {
			ptr := spec.create(t, T)
			require.Nil(t, spec.storageGet(t).Create(spec.context(), ptr))
			id, _ := resources.LookupID(ptr)
			t.Defer(spec.storageGet(t).DeleteByID, spec.context(), T, id)
			return ptr
		})

		s.Test(`an update to the storage should invalidate the local cache unit entity state`, func(t *testcase.T) {
			v := t.I(`value`)
			id, _ := resources.LookupID(v)

			// cache
			_, _ = spec.cacheGet(t).FindByID(spec.context(), spec.new(T), id)   // should trigger caching
			_, _ = iterators.Count(spec.cacheGet(t).FindAll(spec.context(), T)) // should trigger caching

			// mutate
			vUpdated := spec.create(t, T)
			require.Nil(t, resources.SetID(vUpdated, id))
			require.Nil(t, spec.cacheGet(t).Update(spec.context(), vUpdated))
			waiter.Wait()

			ptr := spec.new(T)
			found, err := spec.cacheGet(t).FindByID(spec.context(), ptr, id) // should trigger caching
			require.Nil(t, err)
			require.True(t, found)
			require.Equal(t, vUpdated, ptr)
		})

		s.Test(`a delete by id to the storage should invalidate the local cache unit entity state`, func(t *testcase.T) {
			v := t.I(`value`)
			id, _ := resources.LookupID(v)

			// cache
			_, _ = spec.cacheGet(t).FindByID(spec.context(), spec.new(T), id)   // should trigger caching
			_, _ = iterators.Count(spec.cacheGet(t).FindAll(spec.context(), T)) // should trigger caching

			// delete
			require.Nil(t, spec.cacheGet(t).DeleteByID(spec.context(), T, id))

			async.Assert(t, func(tb testing.TB) {
				found, err := spec.cacheGet(t).FindByID(spec.context(), spec.new(T), id)
				require.Nil(tb, err)
				require.False(tb, found)
			})
		})

		s.Test(`a delete all entity in the storage should invalidate the local cache unit entity state`, func(t *testcase.T) {
			v := t.I(`value`)
			id, _ := resources.LookupID(v)

			// cache
			_, _ = spec.cacheGet(t).FindByID(spec.context(), spec.new(T), id)   // should trigger caching
			_, _ = iterators.Count(spec.cacheGet(t).FindAll(spec.context(), T)) // should trigger caching

			// delete
			require.Nil(t, spec.cacheGet(t).DeleteAll(spec.context(), T))
			waiter.Wait()

			found, err := spec.cacheGet(t).FindByID(spec.context(), spec.new(T), id) // should trigger caching
			require.Nil(t, err)
			require.False(t, found)
		})
	})
}

func (spec Cache) cache() testcase.Var {
	return testcase.Var{
		Name: `cache`,
		Init: func(t *testcase.T) interface{} {
			return spec.NewCache(t, spec.storageGet(t))
		},
	}
}

func (spec Cache) cacheGet(t *testcase.T) toggler.Storage {
	return spec.cache().Get(t).(toggler.Storage)
}

func (spec Cache) storage() testcase.Var {
	return testcase.Var{
		Name: `storage`,
		Init: func(t *testcase.T) interface{} /* [toggler.Storage] */ {
			return sh.StorageGet(t)
		},
	}
}

func (spec Cache) storageGet(t *testcase.T) toggler.Storage {
	return spec.storage().Get(t).(toggler.Storage)
}

type togglerStorageStub struct {
	toggler.Storage
	count struct {
		FindByID int
	}
}

func (stub *togglerStorageStub) FindByID(ctx context.Context, ptr, id interface{}) (_found bool, _err error) {
	stub.count.FindByID++
	return stub.Storage.FindByID(ctx, ptr, id)
}

func (spec Cache) expectResultCachingFor(s *testcase.Spec, T interface{}) {
	s.Context(reflects.SymbolicName(T), func(s *testcase.Spec) {
		value := s.Let(`value`, func(t *testcase.T) interface{} {
			ptr := spec.create(t, T)
			storage := spec.storageGet(t)
			require.Nil(t, storage.Create(spec.context(), ptr))
			id, _ := resources.LookupID(ptr)
			t.Defer(storage.DeleteByID, spec.context(), T, id)
			return ptr
		})

		s.Then(`it will return the value`, func(t *testcase.T) {
			v := spec.new(T)
			id, found := resources.LookupID(value.Get(t))
			require.True(t, found)
			found, err := spec.cacheGet(t).FindByID(spec.context(), v, id)
			require.Nil(t, err)
			require.True(t, found)
			require.Equal(t, value.Get(t), v)
		})

		s.And(`after value already cached`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				id, found := resources.LookupID(value.Get(t))
				require.True(t, found)
				v := contracts2.IsFindable(t, T, spec.storageGet(t), spec.context(), id)
				require.Equal(t, value.Get(t), v)
			})

			s.And(`value is suddenly updated `, func(s *testcase.Spec) {
				valueWithNewContent := s.Let(`value-with-new-content`, func(t *testcase.T) interface{} {
					id, found := resources.LookupID(value.Get(t))
					require.True(t, found)
					nv := spec.create(t, T)
					require.Nil(t, resources.SetID(nv, id))
					return nv
				})

				s.Before(func(t *testcase.T) {
					ptr := valueWithNewContent.Get(t)
					require.Nil(t, spec.cacheGet(t).Update(spec.context(), ptr))
					waiter.Wait()
				})

				s.Then(`it will return the new value instead the old one`, func(t *testcase.T) {
					id, found := resources.LookupID(value.Get(t))
					require.True(t, found)
					require.NotEmpty(t, id)
					async.Assert(t, func(tb testing.TB) {
						v := spec.new(T)
						found, err := spec.cacheGet(t).FindByID(spec.context(), v, id)
						require.Nil(tb, err)
						require.True(tb, found)
						tb.Log(`actually`, v)
						require.Equal(tb, valueWithNewContent.Get(t), v)
					})
				})
			})
		})

		s.And(`on multiple request`, func(s *testcase.Spec) {
			s.Then(`it will return it consistently`, func(t *testcase.T) {
				value := value.Get(t)
				id, found := resources.LookupID(value)
				require.True(t, found)

				for i := 0; i < 42; i++ {
					v := spec.new(T)
					found, err := spec.cacheGet(t).FindByID(spec.context(), v, id)
					require.Nil(t, err)
					require.True(t, found)
					require.Equal(t, value, v)
				}
			})

			s.When(`the storage is sensitive to continuous requests`, func(s *testcase.Spec) {
				stub := s.Let(`stub`, func(t *testcase.T) interface{} {
					return &togglerStorageStub{Storage: sh.StorageGet(t)}
				})
				stubGet := func(t *testcase.T) *togglerStorageStub { return stub.Get(t).(*togglerStorageStub) }

				spec.storage().Let(s, func(t *testcase.T) interface{} {
					return stubGet(t)
				})

				s.Then(`it will only bother the storage for the value once`, func(t *testcase.T) {
					var nv interface{}
					value := value.Get(t)
					id, found := resources.LookupID(value)
					require.True(t, found)

					nv = spec.new(T)
					// trigger caching
					found, err := spec.cacheGet(t).FindByID(spec.context(), nv, id)
					require.Nil(t, err)
					require.True(t, found)
					require.Equal(t, value, nv)
					require.Equal(t, 1, stubGet(t).count.FindByID)

					waiter.Wait()

					nv = spec.new(T)
					// should use caching else trigger mock error
					found, err = spec.cacheGet(t).FindByID(spec.context(), nv, id)
					require.Nil(t, err)
					require.True(t, found)
					require.Equal(t, value, nv)
					require.Equal(t, 1, stubGet(t).count.FindByID)
				})
			})
		})
	})
}

func (spec Cache) context() context.Context {
	return spec.FixtureFactory.Context()
}

func (spec Cache) new(T interface{}) interface{} {
	return reflect.New(reflect.TypeOf(T)).Interface()
}

func (spec Cache) create(t *testcase.T, T interface{}) interface{} {
	return spec.FixtureFactory.Dynamic(t).Create(T)
}
