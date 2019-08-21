package specs

import (
	"testing"
	"time"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/toggler/services/security"
	. "github.com/adamluzsi/toggler/testing"
	"github.com/stretchr/testify/require"
)

type TokenFinderSpec struct {
	Subject interface {
		security.TokenFinder
		specs.MinimumRequirements
	}

	specs.FixtureFactory
}

func (spec TokenFinderSpec) Test(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Let(`uid`, func(t *testcase.T) interface{} { return ExampleUniqUserID() })
	s.Let(`token object`, func(t *testcase.T) interface{} {
		return &security.Token{
			OwnerUID: t.I(`uid`).(string),
			SHA512:   t.I(`token SHA512`).(string),
			IssuedAt: fixtures.RandomTimeUTC(),
			Duration: 1 * time.Second,
		}
	})

	s.Before(func(t *testcase.T) {
		require.Nil(t, spec.Subject.Truncate(spec.Context(), security.Token{}))
	})

	s.After(func(t *testcase.T) {
		require.Nil(t, spec.Subject.Truncate(spec.Context(), security.Token{}))
	})

	s.Describe(`FindTokenBySHA512Hex`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (*security.Token, error) {
			return spec.Subject.FindTokenBySHA512Hex(spec.Context(), t.I(`token SHA512`).(string))
		}

		s.Let(`token SHA512`, func(t *testcase.T) interface{} { return `the answer is 42` })

		s.When(`no token stored in the storage yet`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { require.Nil(t, spec.Subject.Truncate(spec.Context(), security.Token{})) })

			s.Then(`it will return nil token without any error`, func(t *testcase.T) {
				token, err := subject(t)
				require.Nil(t, err)
				require.Nil(t, token)
			})
		})

		s.When(`token is stored in the storage already`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, spec.Subject.Save(spec.Context(), t.I(`token object`).(*security.Token)))
			})

			s.Then(`token will be retrieved`, func(t *testcase.T) {
				actual, err := subject(t)
				expected := t.I(`token object`).(*security.Token)

				require.Nil(t, err)
				require.NotNil(t, actual)
				require.Equal(t, expected, actual)
			})
		})

	})
}

func (spec TokenFinderSpec) Benchmark(b *testing.B) {
	b.Run(`TokenFinderSpec`, func(b *testing.B) {
		b.Skip()
	})
}
