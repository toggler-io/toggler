package contracts

import (
	"context"
	"testing"

	"github.com/toggler-io/toggler/domains/release"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"

	sh "github.com/toggler-io/toggler/spechelper"
)

type EnvironmentStorage struct {
	Subject        func(testing.TB) release.Storage
	Context        func(testing.TB) context.Context
	FixtureFactory func(testing.TB) frameless.FixtureFactory
}

func (c EnvironmentStorage) storage() testcase.Var {
	return testcase.Var{
		Name: "release XY storage",
		Init: func(t *testcase.T) interface{} {
			return c.Subject(t)
		},
	}
}

func (c EnvironmentStorage) storageGet(t *testcase.T) release.Storage {
	return c.storage().Get(t).(release.Storage)
}

func (c EnvironmentStorage) String() string {
	return "EnvironmentStorage"
}

func (c EnvironmentStorage) Test(t *testing.T) {
	c.Spec(testcase.NewSpec(t))
}

func (c EnvironmentStorage) Benchmark(b *testing.B) {
	c.Spec(testcase.NewSpec(b))
}

func (c EnvironmentStorage) Spec(s *testcase.Spec) {
	T := release.Environment{}
	getEnvironmentStorage := func(tb testing.TB) release.EnvironmentStorage {
		return c.Subject(tb).ReleaseEnvironment(c.Context(tb))
	}

	testcase.RunContract(s,
		contracts.Creator{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return getEnvironmentStorage(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Finder{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return getEnvironmentStorage(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Updater{T: T,
			Subject: func(tb testing.TB) contracts.UpdaterSubject {
				return getEnvironmentStorage(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Deleter{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return getEnvironmentStorage(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Publisher{T: T,
			Subject: func(tb testing.TB) contracts.PublisherSubject {
				return getEnvironmentStorage(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		contracts.OnePhaseCommitProtocol{T: release.Environment{},
			Subject: func(tb testing.TB) (frameless.OnePhaseCommitProtocol, contracts.CRD) {
				storage := c.Subject(tb)
				return storage, storage.ReleaseEnvironment(c.Context(tb))
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
	)

	s.Describe(`.FindDeploymentEnvironmentByAlias`, c.specFindDeploymentEnvironmentByAlias)
}

func (c EnvironmentStorage) specFindDeploymentEnvironmentByAlias(s *testcase.Spec) {
	sh.FixtureFactoryLet(s, c.FixtureFactory)
	var (
		env     = s.Let(`env`, func(t *testcase.T) interface{} { return &release.Environment{} })
		alias   = testcase.Var{Name: `alias`}
		subject = func(t *testcase.T) (bool, error) {
			return c.storageGet(t).ReleaseEnvironment(c.Context(t)).FindByAlias(
				c.Context(t),
				alias.Get(t).(string),
				env.Get(t).(*release.Environment),
			)
		}
	)

	testcase.RunContract(s, contracts.FindOne{T: release.Environment{},
		Subject: func(tb testing.TB) contracts.CRD {
			return c.Subject(tb).ReleaseEnvironment(c.Context(tb))
		},
		Context:        c.Context,
		FixtureFactory: c.FixtureFactory,
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
			ctx := c.Context(t)
			contracts.DeleteAllEntity(t, c.storageGet(t).ReleaseEnvironment(ctx), ctx)
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
			ff := sh.FixtureFactoryGet(t)
			ctx := c.Context(t)
			env := ff.Fixture(release.Environment{}, ctx).(release.Environment)
			contracts.CreateEntity(t, c.storageGet(t).ReleaseEnvironment(ctx), ctx, &env)
			return &env
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
