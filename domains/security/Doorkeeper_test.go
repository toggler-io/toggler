package security_test

import (
	"context"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/domains/security"
	. "github.com/toggler-io/toggler/testing"
	"github.com/stretchr/testify/require"
)

func TestDoorkeeper(t *testing.T) {
	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	s.Parallel()

	s.Let(`doorkeeper`, func(t *testcase.T) interface{} {
		return security.NewDoorkeeper(GetStorage(t))
	})

	s.Let(`Token`, func(t *testcase.T) interface{} {
		issuer := security.NewIssuer(GetStorage(t))
		textToken, token, err := issuer.CreateNewToken(context.TODO(), GetUniqUserID(t), nil, nil)
		*(t.I(`*TextToken`).(*string)) = textToken
		require.Nil(t, err)
		return token
	})

	s.Let(`*TextToken`, func(t *testcase.T) interface{} {
		var textToken string
		return &textToken
	})

	s.Let(`TextToken`, func(t *testcase.T) interface{} {
		t.I(`Token`)
		return *(t.I(`*TextToken`).(*string))
	})

	SpecDoorkeeperVerifyTextToken(s)
}

func SpecDoorkeeperVerifyTextToken(s *testcase.Spec) {
	s.Describe(`VerifyTextToken`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (bool, error) {
			return doorkeeper(t).VerifyTextToken(context.TODO(), GetTextToken(t))
		}

		onSuccess := func(t *testcase.T) bool {
			accepted, err := subject(t)
			require.Nil(t, err)
			return accepted
		}

		s.When(`token is a known resource`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				persisted, err := GetStorage(t).FindByID(context.Background(), &security.Token{}, GetToken(t).ID)
				require.Nil(t, err)
				require.True(t, persisted)
			})

			s.Then(`it will verify and accept it`, func(t *testcase.T) {
				require.True(t, onSuccess(t))
			})
		})

		s.When(`token is unknown`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, GetStorage(t).DeleteByID(context.Background(), security.Token{}, GetToken(t).ID))
			})

			s.Then(`it will reject it`, func(t *testcase.T) {
				require.False(t, onSuccess(t))
			})
		})

	})
}

func doorkeeper(t *testcase.T) *security.Doorkeeper {
	return t.I(`doorkeeper`).(*security.Doorkeeper)
}
