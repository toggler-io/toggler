package contracts

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/frameless/resources/contracts"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"github.com/toggler-io/toggler/external/resource/caches"
	"github.com/toggler-io/toggler/external/resource/storages"

	"testing"
)

type Storage struct {
	Subject        func(testing.TB) caches.Storage
	FixtureFactory contracts.FixtureFactory
	EntityTypes    []interface{}
}

func (spec Storage) storage() testcase.Var {
	return testcase.Var{
		Name: "cache storage",
		Init: func(t *testcase.T) interface{} {
			return spec.Subject(t)
		},
	}
}

func (spec Storage) storageGet(t *testcase.T) caches.Storage {
	return spec.storage().Get(t).(caches.Storage)
}

func (spec Storage) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec Storage) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec Storage) Spec(tb testing.TB) {
	s := testcase.NewSpec(tb)
	defer s.Finish()

	once := &sync.Once{}
	s.Before(func(t *testcase.T) {
		once.Do(func() {
			contracts.DeleteAllEntity(t, spec.storageGet(t), spec.FixtureFactory.Context(), caches.CachedQuery{})
		})
	})

	s.Describe(`CachedQuery`, func(s *testcase.Spec) {
		spec.SpecQueryRecord(s)
		s.Describe(`.UpsertMany`, spec.specCachedQueryUpsertMany(caches.CachedQuery{}))
		s.Describe(`.FindByIDs`, spec.specCachedQueryFindManyByIDs(caches.CachedQuery{}))
	})

	for _, T := range spec.EntityTypes {
		T := T
		s.Describe(reflects.SymbolicName(T), func(s *testcase.Spec) {
			s.Describe(`.UpsertMany`, spec.specCachedQueryUpsertMany(T))
			s.Describe(`.FindByIDs`, spec.specCachedQueryFindManyByIDs(T))
		})
	}
}

func (spec Storage) SpecQueryRecord(s *testcase.Spec) {
	T := caches.CachedQuery{}
	testcase.RunContract(s,
		contracts.Creator{T: T,
			FixtureFactory: spec.FixtureFactory,
			Subject:        func(tb testing.TB) contracts.CRD { return spec.Subject(tb) },
		},
		contracts.Finder{T: T,
			FixtureFactory: spec.FixtureFactory,
			Subject:        func(tb testing.TB) contracts.CRD { return spec.Subject(tb) },
		},
		contracts.Updater{T: T,
			FixtureFactory: spec.FixtureFactory,
			Subject:        func(tb testing.TB) contracts.UpdaterSubject { return spec.Subject(tb) },
		},
		contracts.Deleter{T: T,
			FixtureFactory: spec.FixtureFactory,
			Subject:        func(tb testing.TB) contracts.CRD { return spec.Subject(tb) },
		},
	)
}

func (spec Storage) getCount(tb testing.TB, i iterators.Interface) int {
	count, err := iterators.Count(i)
	require.Nil(tb, err)
	return count
}

func (spec Storage) ctxGet(t *testcase.T) context.Context {
	return testcase.Var{
		Name: "cache storage request context",
		Init: func(t *testcase.T) interface{} {
			return context.Background()
		},
	}.Get(t).(context.Context)
}

func (spec Storage) specCachedQueryUpsertMany(T resources.T) func(s *testcase.Spec) {
	rT := reflect.TypeOf(T)
	return func(s *testcase.Spec) {
		var (
			ents    = testcase.Var{Name: `entities`}
			entsGet = func(t *testcase.T) []interface{} { return ents.Get(t).([]interface{}) }
			subject = func(t *testcase.T) error {
				return spec.storageGet(t).UpsertMany(spec.ctxGet(t), entsGet(t)...)
			}
		)

		var (
			ent1 = s.Let(`entity 1`, func(t *testcase.T) interface{} {
				return spec.FixtureFactory.Create(T)
			})
			ent2 = s.Let(`entity 2`, func(t *testcase.T) interface{} {
				return spec.FixtureFactory.Create(T)
			})
		)

		s.After(func(t *testcase.T) {
			id, ok := resources.LookupID(ent1.Get(t))
			if !ok {
				return
			}
			_ = spec.storageGet(t).DeleteByID(spec.ctxGet(t), T, id)
		})
		s.After(func(t *testcase.T) {
			id, ok := resources.LookupID(ent2.Get(t))
			if !ok {
				return
			}
			_ = spec.storageGet(t).DeleteByID(spec.ctxGet(t), T, id)
		})

		s.When(`entities absent from the storage`, func(s *testcase.Spec) {
			ents.Let(s, func(t *testcase.T) interface{} {
				return []interface{}{ent1.Get(t), ent2.Get(t)}
			})

			s.Then(`they will be saved`, func(t *testcase.T) {
				require.Nil(t, subject(t))

				ent1ID, ok := resources.LookupID(ent1.Get(t))
				require.True(t, ok, `entity 1 should have id`)
				actual1 := reflect.New(rT).Interface()
				found, err := spec.storageGet(t).FindByID(spec.ctxGet(t), actual1, ent1ID)
				require.Nil(t, err)
				require.True(t, found, `entity 1 was expected to be stored`)
				require.Equal(t, ent1.Get(t), actual1)

				ent2ID, ok := resources.LookupID(ent2.Get(t))
				require.True(t, ok, `entity 2 should have id`)
				actual2 := reflect.New(rT).Interface()
				found, err = spec.storageGet(t).FindByID(spec.ctxGet(t), actual2, ent2ID)
				require.Nil(t, err)
				require.True(t, found, `entity 2 was expected to be stored`)
				require.Equal(t, ent2.Get(t), actual2)
			})

			s.And(`entities already have a storage string id`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, storages.EnsureID(ent1.Get(t)))
					require.Nil(t, storages.EnsureID(ent2.Get(t)))
				})

				s.Then(`they will be saved`, func(t *testcase.T) {
					require.Nil(t, subject(t))

					ent1ID, ok := resources.LookupID(ent1.Get(t))
					require.True(t, ok, `entity 1 should have id`)
					actual1 := reflect.New(rT).Interface()
					found, err := spec.storageGet(t).FindByID(spec.ctxGet(t), actual1, ent1ID)
					require.Nil(t, err)
					require.True(t, found, `entity 1 was expected to be stored`)
					require.Equal(t, ent1.Get(t), actual1)

					ent2ID, ok := resources.LookupID(ent2.Get(t))
					require.True(t, ok, `entity 2 should have id`)
					found, err = spec.storageGet(t).FindByID(spec.ctxGet(t), reflect.New(rT).Interface(), ent2ID)
					require.Nil(t, err)
					require.True(t, found, `entity 2 was expected to be stored`)
				})
			})
		})

		s.When(`entities present in the storage`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				contracts.CreateEntity(t, spec.storageGet(t), spec.ctxGet(t), ent1.Get(t))
				contracts.CreateEntity(t, spec.storageGet(t), spec.ctxGet(t), ent2.Get(t))
			})

			ents.Let(s, func(t *testcase.T) interface{} {
				return []interface{}{ent1.Get(t), ent2.Get(t)}
			})

			s.Then(`they will be saved`, func(t *testcase.T) {
				require.Nil(t, subject(t))

				ent1ID, ok := resources.LookupID(ent1.Get(t))
				require.True(t, ok, `entity 1 should have id`)
				found, err := spec.storageGet(t).FindByID(spec.ctxGet(t), &caches.CachedQuery{}, ent1ID)
				require.Nil(t, err)
				require.True(t, found, `entity 1 was expected to be stored`)

				ent2ID, ok := resources.LookupID(ent2.Get(t))
				require.True(t, ok, `entity 2 should have id`)
				found, err = spec.storageGet(t).FindByID(spec.ctxGet(t), &caches.CachedQuery{}, ent2ID)
				require.Nil(t, err)
				require.True(t, found, `entity 2 was expected to be stored`)
			})

			s.Then(`total count of the entities will not increase`, func(t *testcase.T) {
				require.Nil(t, subject(t))
				count, err := iterators.Count(spec.storageGet(t).FindAll(spec.ctxGet(t), caches.CachedQuery{}))
				require.Nil(t, err)
				require.Equal(t, len(entsGet(t)), count)
			})

			s.And(`at least one of the entity that being upserted has updated content`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					t.Log(`and entity 1 has updated content`)
					id := spec.getID(t, ent1.Get(t))
					n := spec.FixtureFactory.Create(T)
					require.Nil(t, resources.SetID(n, id))
					ent1.Set(t, n)
				})

				s.Then(`the updated data will be saved`, func(t *testcase.T) {
					require.Nil(t, subject(t))

					ent1ID, ok := resources.LookupID(ent1.Get(t))
					require.True(t, ok, `entity 1 should have id`)
					actual := &caches.CachedQuery{}
					found, err := spec.storageGet(t).FindByID(spec.ctxGet(t), actual, ent1ID)
					require.Nil(t, err)
					require.True(t, found, `entity 1 was expected to be stored`)
					require.Equal(t, ent1.Get(t), actual)

					ent2ID, ok := resources.LookupID(ent2.Get(t))
					require.True(t, ok, `entity 2 should have id`)
					found, err = spec.storageGet(t).FindByID(spec.ctxGet(t), &caches.CachedQuery{}, ent2ID)
					require.Nil(t, err)
					require.True(t, found, `entity 2 was expected to be stored`)
				})

				s.Then(`total count of the entities will not increase`, func(t *testcase.T) {
					require.Nil(t, subject(t))
					count, err := iterators.Count(spec.storageGet(t).FindAll(spec.ctxGet(t), caches.CachedQuery{}))
					require.Nil(t, err)
					require.Equal(t, len(entsGet(t)), count)
				})
			})
		})
	}
}

func (spec Storage) specCachedQueryFindManyByIDs(T resources.T) func(s *testcase.Spec) {
	rT := reflect.TypeOf(T)
	MakeTSlice := func() interface{} {
		return reflect.MakeSlice(reflect.SliceOf(rT), 0, 0).Interface()
	}
	Append := func(slice interface{}, values ...interface{}) interface{} {
		var vs []reflect.Value
		for _, v := range values {
			vs = append(vs, reflect.ValueOf(v))
		}
		return reflect.Append(reflect.ValueOf(slice), vs...).Interface()
	}
	bv := func(v interface{}) interface{} { return reflects.BaseValueOf(v).Interface() }
	return func(s *testcase.Spec) {
		var (
			ids     = testcase.Var{Name: `entities ids`}
			idsGet  = func(t *testcase.T) []interface{} { return ids.Get(t).([]interface{}) }
			subject = func(t *testcase.T) iterators.Interface {
				return spec.storageGet(t).FindByIDs(spec.ctxGet(t), T, idsGet(t)...)
			}
		)

		var (
			ent1 = s.Let(`stored entity 1`, func(t *testcase.T) interface{} {
				ptr := spec.FixtureFactory.Create(T)
				contracts.CreateEntity(t, spec.storageGet(t), spec.ctxGet(t), ptr)
				return ptr
			})
			ent2 = s.Let(`stored entity 2`, func(t *testcase.T) interface{} {
				ptr := spec.FixtureFactory.Create(T)
				contracts.CreateEntity(t, spec.storageGet(t), spec.ctxGet(t), ptr)
				return ptr
			})
		)

		s.When(`id list is empty`, func(s *testcase.Spec) {
			ids.Let(s, func(t *testcase.T) interface{} {
				return []interface{}{}
			})

			s.Then(`result is an empty list`, func(t *testcase.T) {
				count, err := iterators.Count(subject(t))
				require.Nil(t, err)
				require.Equal(t, 0, count)
			})
		})

		s.When(`id list contains ids stored in the storage`, func(s *testcase.Spec) {
			ids.Let(s, func(t *testcase.T) interface{} {
				return []interface{}{spec.getID(t, ent1.Get(t)), spec.getID(t, ent2.Get(t))}
			})

			s.Then(`it will return all entities`, func(t *testcase.T) {
				actual2 := MakeTSlice()
				require.Nil(t, iterators.Collect(spec.storageGet(t).FindAll(spec.ctxGet(t), T), &actual2))

				expected := Append(MakeTSlice(), bv(ent1.Get(t)), bv(ent2.Get(t)))
				actual := MakeTSlice()

				require.Nil(t, iterators.Collect(subject(t), &actual))
				require.ElementsMatch(t, expected, actual)
			})
		})

		s.When(`id list contains at least one id that doesn't have stored entity`, func(s *testcase.Spec) {
			ids.Let(s, func(t *testcase.T) interface{} {
				return []interface{}{spec.getID(t, ent1.Get(t)), spec.getID(t, ent2.Get(t))}
			})

			s.Before(func(t *testcase.T) {
				contracts.DeleteEntity(t, spec.storageGet(t), spec.ctxGet(t), ent1.Get(t))
			})

			s.Then(`it will eventually yield error`, func(t *testcase.T) {
				list := MakeTSlice()
				require.Error(t, iterators.Collect(subject(t), &list))
			})
		})
	}
}

func (spec Storage) getID(tb testing.TB, ent interface{}) interface{} {
	id, ok := resources.LookupID(ent)
	require.True(tb, ok, `id was expected to be present for the entity`+fmt.Sprintf(` (%#v)`, ent))
	return id
}
