package usecases_test

import (
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUseCases_SetFeatureFlagRolloutStrategyToUseDecisionLogicAPI(t *testing.T) {
	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	SetupSpec(s)
	s.Parallel()

	subject := func(t *testcase.T) error {
		return GetUseCases(t).SetFeatureFlag(
			t.I(`TokenString`).(string),
			GetFeatureFlag(t),
		)
	}

	s.When(`caller have invalid token`, func(s *testcase.Spec) {
		s.Let(`TokenString`, func(t *testcase.T) interface{} { return `invalid token` })

		s.Then(`it will be rejected`, func(t *testcase.T) {
			require.Equal(t, usecases.ErrInvalidToken, subject(t))
		})
	})

	s.When(`caller have a valid token`, func(s *testcase.Spec) {
		s.Let(`TokenString`, func(t *testcase.T) interface{} {
			return CreateToken(t, `rollout manager uniq id`).Token
		})

		s.And(`with valid values`, func(s *testcase.Spec) {
			s.Then(`it will be persisted`, func(t *testcase.T) {
				require.Nil(t, subject(t))

				require.Equal(t, GetFeatureFlag(t), FindStoredFeatureFlag(t))
			})
		})

	})

}
