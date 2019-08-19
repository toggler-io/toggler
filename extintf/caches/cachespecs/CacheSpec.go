package cachespecs

import (
	"context"
	"fmt"
	"testing"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources"
	frmlspecs "github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/toggler/extintf/caches"
	"github.com/adamluzsi/toggler/extintf/storages/inmemory"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
	"github.com/adamluzsi/toggler/usecases"
	"github.com/adamluzsi/toggler/usecases/specs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -source ../../../usecases/Storage.go -destination MockStorage.go -package cachespecs

type CacheSpec struct {
	Factory func(usecases.Storage) caches.Interface

	FixtureFactory interface {
		frmlspecs.FixtureFactory
		SetPilotFeatureFlagID(ffID string) func()
	}
}

func (spec CacheSpec) Test(t *testing.T) {
	s := testcase.NewSpec(t)
	spec.setup(s)

	s.Around(func(t *testcase.T) func() {
		cleanup := func() {
			require.Nil(t, spec.cache(t).Truncate(spec.ctx(rollouts.FeatureFlag{}), rollouts.FeatureFlag{}))
			require.Nil(t, spec.cache(t).Truncate(spec.ctx(rollouts.Pilot{}), rollouts.Pilot{}))
		}

		cleanup()
		return func() { cleanup() }
	})

	s.After(func(t *testcase.T) {
		require.Nil(t, spec.cache(t).Close())
	})

	s.Describe(`CacheSpec`, func(s *testcase.Spec) {
		s.Test(`cache mimics the storage behavior by proxying between storage and the caller`, func(t *testcase.T) {
			specs.StorageSpec{Subject: spec.cache(t), FixtureFactory: spec.FixtureFactory}.Test(t.T)
		})

		SharedSpecForEntityCachingBehavior := func(s *testcase.Spec, T interface{}) {
			title := fmt.Sprintf(`%s cache rules`, reflects.FullyQualifiedName(T))
			s.Describe(title, func(s *testcase.Spec) {

				s.Let(`value`, func(t *testcase.T) interface{} {
					return spec.FixtureFactory.Create(T)
				})

				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.storage(t).Save(spec.ctx(t.I(`value`)), t.I(`value`)))
					_, found := resources.LookupID(t.I(`value`))
					require.True(t, found)
				})

				s.Then(`it will return the value`, func(t *testcase.T) {
					v := reflects.New(T)
					id, found := resources.LookupID(t.I(`value`))
					require.True(t, found)
					found, err := spec.cache(t).FindByID(spec.ctx(v), v, id)
					require.Nil(t, err)
					require.True(t, found)
					require.Equal(t, t.I(`value`), v)
				})

				s.And(`after value already cached`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						v := reflects.New(T)
						id, found := resources.LookupID(t.I(`value`))
						require.True(t, found)
						found, err := spec.cache(t).FindByID(spec.ctx(v), v, id)
						require.Nil(t, err)
						require.True(t, found)
						require.Equal(t, t.I(`value`), v)
					})

					s.And(`value is suddenly updated `, func(s *testcase.Spec) {
						s.Let(`value-with-new-content`, func(t *testcase.T) interface{} {
							id, found := resources.LookupID(t.I(`value`))
							require.True(t, found)
							nv := spec.FixtureFactory.Create(T)
							require.Nil(t, resources.SetID(nv, id))
							return nv
						})

						s.Before(func(t *testcase.T) {
							v := t.I(`value-with-new-content`)
							require.Nil(t, spec.cache(t).Update(spec.ctx(v), v))
						})

						s.Then(`it will return the new value instead the old one`, func(t *testcase.T) {
							v := reflects.New(T)
							id, found := resources.LookupID(t.I(`value`))
							require.True(t, found)
							found, err := spec.cache(t).FindByID(spec.ctx(v), v, id)
							require.Nil(t, err)
							require.True(t, found)
							require.Equal(t, t.I(`value-with-new-content`), v)
						})
					})
				})

				s.And(`on multiple request`, func(s *testcase.Spec) {
					s.Then(`it will return it consistently`, func(t *testcase.T) {
						value := t.I(`value`)
						id, found := resources.LookupID(value)
						require.True(t, found)

						for i := 0; i < 42; i++ {
							v := reflects.New(T)
							found, err := spec.cache(t).FindByID(spec.ctx(v), v, id)
							require.Nil(t, err)
							require.True(t, found)
							require.Equal(t, value, v)
						}
					})

					s.And(`when the storage is sensitive to continuous requests`, func(s *testcase.Spec) {
						s.Context(`for finding the same flag By ID`, func(s *testcase.Spec) {
							spec.mockStorage(s, func(t *testcase.T, storage *MockStorage) {
								storage.EXPECT().Save(gomock.Any(), gomock.Any()).
									AnyTimes().
									DoAndReturn(func(ctx context.Context, e interface{}) error {
										return resources.SetID(e, fixtures.RandomString(7))
									})

								storage.EXPECT().FindByID(gomock.Any(), gomock.Any(), gomock.Any()).
									Times(1).
									DoAndReturn(func(ctx context.Context, ptr interface{}, ID string) (bool, error) {
										value := t.I(`value`)
										id, found := resources.LookupID(value)
										require.True(t, found)
										require.Equal(t, ID, id)
										require.Nil(t, reflects.Link(value, ptr))
										return true, nil
									})

								storage.EXPECT().Truncate(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

								storage.EXPECT().Close().AnyTimes().Return(nil)
							})

							s.Then(`it will only bother the storage for the value once`, func(t *testcase.T) {
								var nv interface{}
								value := t.I(`value`)
								id, found := resources.LookupID(value)
								require.True(t, found)

								nv = reflects.New(T)
								found, err := spec.cache(t).FindByID(spec.ctx(nv), nv, id)
								require.Nil(t, err)
								require.True(t, found)
								require.Equal(t, value, nv)

								nv = reflects.New(T)
								found, err = spec.cache(t).FindByID(spec.ctx(nv), nv, id)
								require.Nil(t, err)
								require.True(t, found)
								require.Equal(t, value, nv)
							})
						})
					})
				})
			})
		}

		SharedSpecForEntityCachingBehavior(s, rollouts.FeatureFlag{})
		SharedSpecForEntityCachingBehavior(s, rollouts.Pilot{})
		SharedSpecForEntityCachingBehavior(s, security.Token{})

	})
}

func (spec CacheSpec) setup(s *testcase.Spec) {
	s.Let(`cache`, func(t *testcase.T) interface{} {
		return spec.Factory(spec.storage(t))
	})

	s.Let(`storage`, func(t *testcase.T) interface{} {
		return inmemory.New()
	})
}

func (spec CacheSpec) cache(t *testcase.T) caches.Interface {
	return t.I(`cache`).(caches.Interface)
}

func (spec CacheSpec) storage(t *testcase.T) usecases.Storage {
	return t.I(`storage`).(usecases.Storage)
}

func (spec CacheSpec) ctx(e interface{}) context.Context {
	return spec.FixtureFactory.Context()
}

func (spec CacheSpec) mockStorage(s *testcase.Spec, setupMockBehavior func(*testcase.T, *MockStorage)) {
	s.Let(`storage`, func(t *testcase.T) interface{} {
		mock := NewMockStorage(t.I(`storage-ctrl`).(*gomock.Controller))
		setupMockBehavior(t, mock)
		return mock
	})

	s.Let(`storage-ctrl`, func(t *testcase.T) interface{} {
		return gomock.NewController(t.T)
	})

	s.After(func(t *testcase.T) {
		t.I(`storage-ctrl`).(*gomock.Controller).Finish()
	})
}
