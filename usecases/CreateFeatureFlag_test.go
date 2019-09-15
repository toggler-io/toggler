package usecases_test

import (
	"context"
	. "github.com/toggler-io/toggler/testing"
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
		return GetProtectedUsecases(t).CreateFeatureFlag(context.TODO(), GetReleaseFlag(t))
	}

	s.When(`with valid values`, func(s *testcase.Spec) {
		s.Then(`it will be set/persisted`, func(t *testcase.T) {
			require.Nil(t, subject(t))

			require.Equal(t, GetReleaseFlag(t), FindStoredReleaseFlagByName(t))
		})
	})

}
