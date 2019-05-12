package specs

import (
	"testing"

	"github.com/adamluzsi/FeatureFlags/services/security"
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
)

type TokenFinderSpec struct {
	Subject interface {
		security.TokenFinder

		specs.MinimumRequirements
	}
}

func (spec TokenFinderSpec) Test(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Let(`uid`, func(t *testcase.T) interface{} { return ExampleUniqUserID() })
	s.Let(`token object`, func(t *testcase.T) interface{} {
		return &security.Token{
			UserUID: t.I(`uid`).(string),
			Token:   t.I(`token string`).(string),
		}
	})

	s.Before(func(t *testcase.T) {
		require.Nil(t, spec.Subject.Truncate(security.Token{}))
	})

	s.Describe(`FindTokenByTokenString`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (*security.Token, error) {
			return spec.Subject.FindTokenByTokenString(t.I(`token string`).(string))
		}

		s.Let(`token string`, func(t *testcase.T) interface{} { return `the answer is 42` })

		s.When(`no token stored in the storage yet`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { require.Nil(t, spec.Subject.Truncate(security.Token{})) })

			s.Then(`it will retun nil token without any error`, func(t *testcase.T) {
				token, err := subject(t)
				require.Nil(t, err)
				require.Nil(t, token)
			})
		})

		s.When(`token is stored in the storage already`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, spec.Subject.Save(t.I(`token object`).(*security.Token)))
			})

			s.Then(`token will be retrieved`, func(t *testcase.T) {
				token, err := subject(t)
				require.Nil(t, err)
				require.NotNil(t, token)
				require.Equal(t, t.I(`token object`), token)
			})
		})

	})
}
