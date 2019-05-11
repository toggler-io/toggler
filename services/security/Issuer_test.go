package security_test

import (
	"github.com/adamluzsi/FeatureFlags/services/security"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	. "github.com/adamluzsi/FeatureFlags/testing"
)

func TestIssuer(t *testing.T) {
	t.Skip(`TODO: implement`)

	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)

	issuer := func(t *testcase.T) *security.Issuer {
		return &security.Issuer{Storage: t.I(`TestStorage`).(*TestStorage)}
	}

	s.Describe(`CreateNewToken`, func(s *testcase.Spec) {

		subject := func(t *testcase.T) (*security.Token, error) {
			userUID := t.I(`userUID`).(string)
			issueAt, _ := t.I(`issueAt`).(*time.Time)
			duration, _ := t.I(`duration`).(*time.Duration)

			return issuer(t).CreateNewToken(userUID, issueAt, duration)
		}

		onSuccess := func(t *testcase.T) *security.Token {
			token, err := subject(t)
			require.Nil(t, err)
			require.NotNil(t, token)
			return token
		}

		onFailure := func(t *testcase.T) error {
			token, err := subject(t)
			require.Nil(t, token)
			require.NotNil(t, err)
			return err
		}

		givenWeHaveValidParameters := func(s *testcase.Spec) {
			s.Let(`userUID`, func(t *testcase.T) interface{} { return ExampleUniqUserID() })
			s.Let(`issueAt`, func(t *testcase.T) interface{} { ia := time.Now().UTC(); return &ia })
			s.Let(`duration`, func(t *testcase.T) interface{} { d := 42 * time.Hour; return &d })
		}

		s.When(`all parameter acceptable`, func(s *testcase.Spec) {
			givenWeHaveValidParameters(s)

			s.Then(`we receive a token back`, func(t *testcase.T) {
				token := onSuccess(t)
				require.Equal(t, t.I(`userUID`).(string), token.UserUID)
				require.Equal(t, t.I(`issueAt`).(*time.Time), token.IssuedAt)
				require.Equal(t, t.I(`duration`).(*time.Duration), token.Duration)
			})
		})

		s.When(`userUID is empty`, func(s *testcase.Spec) {
			givenWeHaveValidParameters(s)
			s.Let(`userUID`, func(t *testcase.T) interface{} { return `` })

			s.Then(`we receive error because empty userUID is not acceptable`, func(t *testcase.T) {
				require.Error(t, onFailure(t))
			})
		})

		s.When(`duration is not provided`, func(s *testcase.Spec) {
			givenWeHaveValidParameters(s)
			s.Let(`duration`, func(t *testcase.T) interface{} { return nil })

			s.Then(`it will create a token that can't expire`, func(t *testcase.T) {
				token := onSuccess(t)
				require.False(t, token.IsExpirable())
			})
		})
	})

}
