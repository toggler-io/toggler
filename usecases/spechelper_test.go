package usecases_test

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
)

func SetupSpec(s *testcase.Spec) {
	s.Let(`UseCases`, func(t *testcase.T) interface{} {
		return usecases.NewUseCases(GetStorage(t))
	})
}

func GetRolloutManager(t *testcase.T) *rollouts.RolloutManager {
	rm := rollouts.NewRolloutManager(GetStorage(t))
	return rm
}

func EnsureFlag(t *testcase.T, name string) {
	rm := rollouts.NewRolloutManager(GetStorage(t))
	require.Nil(t, rm.UpdateFeatureFlagRolloutPercentage(name, 0))
}

func GetUseCases(t *testcase.T) *usecases.UseCases {
	return t.I(`UseCases`).(*usecases.UseCases)
}
