package usecases_test

import (
	"context"
	. "github.com/toggler-io/toggler/testing"
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
		return GetProtectedUsecases(t).UpdateFeatureFlag(context.TODO(), GetReleaseFlag(t))
	}

	s.Context(`Given the feature flag already exists`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, GetProtectedUsecases(t).CreateFeatureFlag(context.TODO(), GetReleaseFlag(t)))
		})

		s.Then(`it will be update changes`, func(t *testcase.T) {
			GetReleaseFlag(t).Rollout.Strategy.Percentage = 42
			require.Nil(t, subject(t))
			require.Equal(t, GetReleaseFlag(t), FindStoredReleaseFlagByName(t))
		})
	})

}
