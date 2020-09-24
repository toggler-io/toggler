package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/deployment"

	. "github.com/toggler-io/toggler/testing"
)

type EnvironmentStorage struct {
	Subject interface {
		resources.Creator
		resources.Finder
		resources.Updater
		resources.Deleter
		resources.CreatorPublisher
		resources.UpdaterPublisher
		resources.DeleterPublisher
		resources.OnePhaseCommitProtocol
		deployment.EnvironmentFinder
	}

	FixtureFactory FixtureFactory
}

func (spec EnvironmentStorage) Test(t *testing.T) {
	t.Run(`EnvironmentStorage`, func(t *testing.T) {
		spec.Spec(t)
	})
}

func (spec EnvironmentStorage) Benchmark(b *testing.B) {
	b.Run(`EnvironmentStorage`, func(b *testing.B) {
		spec.Spec(b)
	})
}

func (spec EnvironmentStorage) Spec(tb testing.TB) {
	s := testcase.NewSpec(tb)

	specs.Run(tb,
		specs.Creator{
			Subject:        spec.Subject,
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
		},
		specs.Finder{
			Subject:        spec.Subject,
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
		},
		specs.Updater{
			Subject:        spec.Subject,
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
		},
		specs.Deleter{
			Subject:        spec.Subject,
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
		},
		specs.OnePhaseCommitProtocol{
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
			Subject:        spec.Subject,
		},
		specs.CreatorPublisher{
			Subject:        spec.Subject,
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
		},
		specs.UpdaterPublisher{
			Subject:        spec.Subject,
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
		},
		specs.DeleterPublisher{
			Subject:        spec.Subject,
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
		},
	)

	s.Describe(`FindDeploymentEnvironmentByAlias`, spec.specFindDeploymentEnvironmentByAlias)
}

func (spec EnvironmentStorage) specFindDeploymentEnvironmentByAlias(s *testcase.Spec) {
	var subject = func(t *testcase.T) (bool, error) {
		return spec.Subject.FindDeploymentEnvironmentByAlias(
			spec.FixtureFactory.Context(),
			t.I(`alias`).(string),
			t.I(`env`).(*deployment.Environment),
		)
	}

	s.Let(`env`, func(t *testcase.T) interface{} {
		return &deployment.Environment{}
	})

	s.When(`no environment stored`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, spec.Subject.DeleteAll(spec.FixtureFactory.Context(), deployment.Environment{}))
		})

		s.LetValue(`alias`, `some-fake-value`)

		s.Then(`it yields no result`, func(t *testcase.T) {
			found, err := subject(t)
			require.Nil(t, err)
			require.False(t, found)
		})
	})

	s.When(`environment stored in the system`, func(s *testcase.Spec) {
		s.Let(`stored-env`, func(t *testcase.T) interface{} {
			env := spec.FixtureFactory.Create(deployment.Environment{}).(*deployment.Environment)
			require.Nil(t, spec.Subject.Create(spec.FixtureFactory.Context(), env))
			t.Defer(func() { require.Nil(t, spec.Subject.DeleteByID(spec.FixtureFactory.Context(), *env, env.ID)) })
			return env
		})

		s.And(`alias defined as id`, func(s *testcase.Spec) {
			s.Let(`alias`, func(t *testcase.T) interface{} {
				return t.I(`stored-env`).(*deployment.Environment).ID
			})

			s.Then(`it find the environment value`, func(t *testcase.T) {
				found, err := subject(t)
				require.Nil(t, err)
				require.True(t, found)
				require.Equal(t, t.I(`stored-env`), t.I(`env`))
			})
		})

		s.And(`alias defined as name`, func(s *testcase.Spec) {
			s.Let(`alias`, func(t *testcase.T) interface{} {
				return t.I(`stored-env`).(*deployment.Environment).Name
			})

			s.Then(`it find the environment value`, func(t *testcase.T) {
				found, err := subject(t)
				require.Nil(t, err)
				require.True(t, found)
				require.Equal(t, t.I(`stored-env`), t.I(`env`))
			})
		})
	})
}
