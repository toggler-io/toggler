package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/deployment"
)

type EnvironmentFinderStorage struct {
	Subject deployment.Storage
	specs.FixtureFactory
}

func (spec EnvironmentFinderStorage) Test(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Context(`EnvironmentFinderStorage`, func(s *testcase.Spec) {
		s.Describe(`FindDeploymentEnvironmentByAlias`, func(s *testcase.Spec) {
			var subject = func(t *testcase.T) (bool, error) {
				return spec.Subject.FindDeploymentEnvironmentByAlias(
					spec.Context(),
					t.I(`alias`).(string),
					t.I(`env`).(*deployment.Environment),
				)
			}

			s.Let(`env`, func(t *testcase.T) interface{} {
				return &deployment.Environment{}
			})

			s.When(`no environment stored`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.DeleteAll(spec.Context(), deployment.Environment{}))
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
					env := spec.Create(deployment.Environment{}).(*deployment.Environment)
					require.Nil(t, spec.Subject.Create(spec.Context(), env))
					t.Defer(func() { require.Nil(t, spec.Subject.DeleteByID(spec.Context(), *env, env.ID)) })
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
		})
	})
}

func (spec EnvironmentFinderStorage) Benchmark(b *testing.B) {
	b.Run(`EnvironmentFinderStorage`, func(b *testing.B) {
		b.Skip()
	})
}
