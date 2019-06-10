package usecases_test

import (
	. "github.com/adamluzsi/toggler/testing"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUseCases_CreateFeatureFlagRolloutStrategyToUseDecisionLogicAPI(t *testing.T) {
	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	SetupSpec(s)
	s.Parallel()

	subject := func(t *testcase.T) error {
		return GetProtectedUsecases(t).CreateFeatureFlag(GetFeatureFlag(t))
	}

	s.When(`with valid values`, func(s *testcase.Spec) {
		s.Then(`it will be set/persisted`, func(t *testcase.T) {
			require.Nil(t, subject(t))

			require.Equal(t, GetFeatureFlag(t), FindStoredFeatureFlagByName(t))
		})
	})

}
