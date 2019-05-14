package security_test

import (
	"testing"

	"github.com/adamluzsi/FeatureFlags/services/security"
	testing2 "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
)

func TestDoorkeeper(t *testing.T) {
	s := testcase.NewSpec(t)
	testing2.SetupSpecCommonVariables(s)
	s.Parallel()

	s.Let(`doorkeeper`, func(t *testcase.T) interface{} {
		return security.NewDoorkeeper(testing2.GetStorage(t))
	})

	s.Let(`Token`, func(t *testcase.T) interface{} {
		issuer := security.NewIssuer(testing2.GetStorage(t))
		token, err := issuer.CreateNewToken(testing2.GetUniqUserID(t), nil, nil)
		require.Nil(t, err)
		return token
	})

	SpecDoorkeeperVerifyTokenString(s)
}

func SpecDoorkeeperVerifyTokenString(s *testcase.Spec) {
	s.Describe(`VerifyTokenString`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (bool, error) {
			return doorkeeper(t).VerifyTokenString(GetToken(t).Token)
		}

		onSuccess := func(t *testcase.T) bool {
			accepted, err := subject(t)
			require.Nil(t, err)
			return accepted
		}

		s.When(`token is a known resource`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				persisted, err := testing2.GetStorage(t).FindByID(GetToken(t).ID, &security.Token{})
				require.Nil(t, err)
				require.True(t, persisted)
			})

			s.Then(`it will verify and accept it`, func(t *testcase.T) {
				require.True(t, onSuccess(t))
			})
		})

		s.When(`token is unknown`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, testing2.GetStorage(t).DeleteByID(GetToken(t), GetToken(t).ID))
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

func GetToken(t *testcase.T) *security.Token {
	return t.I(`Token`).(*security.Token)
}
