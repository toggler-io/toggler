package contracts

import (
	"context"
	"github.com/toggler-io/toggler/domains/release"
	"testing"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"

	sh "github.com/toggler-io/toggler/spechelper"
)

type EnvironmentStorage struct {
	Subject        func(testing.TB) release.Storage
	FixtureFactory sh.FixtureFactory
}

func (spec EnvironmentStorage) storage() testcase.Var {
	return testcase.Var{
		Name: "release XY storage",
		Init: func(t *testcase.T) interface{} {
			return spec.Subject(t)
		},
	}
}

func (spec EnvironmentStorage) storageGet(t *testcase.T) release.Storage {
	return spec.storage().Get(t).(release.Storage)
}

func (spec EnvironmentStorage) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec EnvironmentStorage) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec EnvironmentStorage) Spec(tb testing.TB) {
	testcase.NewSpec(tb).Describe(`EnvironmentStorage`, func(s *testcase.Spec) {
		T := release.Environment{}
		getEnvironmentStorage := func(tb testing.TB) release.EnvironmentStorage {
			return spec.Subject(tb).ReleaseEnvironment(spec.FixtureFactory.Context())
		}

		testcase.RunContract(s,
			contracts.Creator{T: T,
				Subject: func(tb testing.TB) contracts.CRD {
					return getEnvironmentStorage(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Finder{T: T,
				Subject: func(tb testing.TB) contracts.CRD {
					return getEnvironmentStorage(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Updater{T: T,
				Subject: func(tb testing.TB) contracts.UpdaterSubject {
					return getEnvironmentStorage(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Deleter{T: T,
				Subject: func(tb testing.TB) contracts.CRD {
					return getEnvironmentStorage(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Publisher{T: T,
				Subject: func(tb testing.TB) contracts.PublisherSubject {
					return getEnvironmentStorage(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.OnePhaseCommitProtocol{T: release.Environment{},
				Subject: func(tb testing.TB) (frameless.OnePhaseCommitProtocol, contracts.CRD) {
					storage := spec.Subject(tb)
					return storage, storage.ReleaseEnvironment(spec.FixtureFactory.Context())
				},
				FixtureFactory: spec.FixtureFactory,
			},
		)

		s.Describe(`.FindDeploymentEnvironmentByAlias`, spec.specFindDeploymentEnvironmentByAlias)
	})
}

func (spec EnvironmentStorage) specFindDeploymentEnvironmentByAlias(s *testcase.Spec) {
	var (
		env     = s.Let(`env`, func(t *testcase.T) interface{} { return &release.Environment{} })
		alias   = testcase.Var{Name: `alias`}
		subject = func(t *testcase.T) (bool, error) {
			return spec.storageGet(t).ReleaseEnvironment(spec.FixtureFactory.Context()).FindByAlias(
				spec.FixtureFactory.Context(),
				alias.Get(t).(string),
				env.Get(t).(*release.Environment),
			)
		}
	)

	testcase.RunContract(s, contracts.FindOne{T: release.Environment{},
		Subject:        func(tb testing.TB) contracts.CRD { return spec.Subject(tb).ReleaseEnvironment(spec.FixtureFactory.Context()) },
		FixtureFactory: spec.FixtureFactory,
		ToQuery: func(tb testing.TB, resource interface{}, ent contracts.T) contracts.QueryOne {
			var (
				storage   = resource.(release.EnvironmentStorage)
				env       = ent.(*release.Environment)
				idOrAlias string
			)
			if fixtures.Random.Bool() {
				tb.Log(`.ID is used as deployment environment alias`)
				idOrAlias = env.ID
			} else {
				tb.Log(`.Name is used as deployment environment alias`)
				idOrAlias = env.Name
			}
			return func(tb testing.TB, ctx context.Context, ptr contracts.T) (found bool, err error) {
				return storage.FindByAlias(ctx, idOrAlias, ptr.(*release.Environment))
			}
		},
	})

	s.When(`no environment stored`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			ctx := spec.FixtureFactory.Context()
			contracts.DeleteAllEntity(t, spec.storageGet(t).ReleaseEnvironment(ctx), ctx)
		})

		alias.LetValue(s, `some-fake-value`)

		s.Then(`it yields no result`, func(t *testcase.T) {
			found, err := subject(t)
			require.Nil(t, err)
			require.False(t, found)
		})
	})

	s.When(`environment stored in the system`, func(s *testcase.Spec) {
		storedEnv := s.Let(`stored-env`, func(t *testcase.T) interface{} {
			env := spec.FixtureFactory.Create(release.Environment{}).(*release.Environment)
			ctx := spec.FixtureFactory.Context()
			contracts.CreateEntity(t, spec.storageGet(t).ReleaseEnvironment(ctx), ctx, env)
			return env
		}).EagerLoading(s)
		storedEnvGet := func(t *testcase.T) *release.Environment {
			return storedEnv.Get(t).(*release.Environment)
		}

		s.And(`alias defined as id`, func(s *testcase.Spec) {
			alias.Let(s, func(t *testcase.T) interface{} {
				return storedEnvGet(t).ID
			})

			s.Then(`it find the environment value`, func(t *testcase.T) {
				found, err := subject(t)
				require.Nil(t, err)
				require.True(t, found)
				require.Equal(t, storedEnv.Get(t), env.Get(t))
			})
		})

		s.And(`alias defined as name`, func(s *testcase.Spec) {
			alias.Let(s, func(t *testcase.T) interface{} {
				return storedEnvGet(t).Name
			})

			s.Then(`it find the environment value`, func(t *testcase.T) {
				found, err := subject(t)
				require.Nil(t, err)
				require.True(t, found)
				require.Equal(t, storedEnv.Get(t), env.Get(t))
			})
		})
	})
}
