package usecases_test

import (
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUseCases_UpdateFeatureFlag(t *testing.T) {
	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	SetupSpec(s)
	s.Parallel()

	subject := func(t *testcase.T) error {
		return GetProtectedUsecases(t).UpdateFeatureFlag(GetFeatureFlag(t))
	}

	s.Context(`Given the feature flag already exists`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, GetProtectedUsecases(t).CreateFeatureFlag(GetFeatureFlag(t)))
		})

		s.Then(`it will be update changes`, func(t *testcase.T) {
			GetFeatureFlag(t).Rollout.Strategy.Percentage = 42
			require.Nil(t, subject(t))
			require.Equal(t, GetFeatureFlag(t), FindStoredFeatureFlagByName(t))
		})
	})

}
