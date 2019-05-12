package testing

import (
	"math/rand"
	"time"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/testcase"
)

func SetupSpecCommonVariables(s *testcase.Spec) {

	s.Let(`FeatureName`, func(t *testcase.T) interface{} {
		return ExampleFeatureName()
	})

	s.Let(`ExternalPilotID`, func(t *testcase.T) interface{} {
		return ExampleExternalPilotID()
	})

	s.Let(`TestStorage`, func(t *testcase.T) interface{} {
		return NewTestStorage()
	})

	s.Let(`PilotEnrollment`, func(t *testcase.T) interface{} {
		return rand.Intn(2) == 0
	})

	s.Let(`Pilot`, func(t *testcase.T) interface{} {
		return &rollouts.Pilot{
			FeatureFlagID: t.I(`FeatureFlag`).(*rollouts.FeatureFlag).ID,
			ExternalID:    t.I(`ExternalPilotID`).(string),
			Enrolled:      t.I(`PilotEnrollment`).(bool),
		}
	})

	s.Let(`RolloutSeedSalt`, func(t *testcase.T) interface{} { return time.Now().Unix() })
	s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return int(0) })
	s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} { return "" })
	s.Let(`FeatureFlag`, func(t *testcase.T) interface{} {
		ff := &rollouts.FeatureFlag{Name: t.I(`FeatureName`).(string)}
		ff.Rollout.RandSeedSalt = t.I(`RolloutSeedSalt`).(int64)
		ff.Rollout.Strategy.Percentage = t.I(`RolloutPercentage`).(int)
		ff.Rollout.Strategy.URL = t.I(`RolloutApiURL`).(string)
		return ff
	})

}

func GetExternalPilotID(t *testcase.T) string {
	return t.I(`ExternalPilotID`).(string)
}

func GetFeatureFlagName(t *testcase.T) string {
	return t.I(`FeatureName`).(string)
}

func GetStorage(t *testcase.T) *TestStorage {
	return t.I(`TestStorage`).(*TestStorage)
}

func GetFeatureFlag(t *testcase.T) *rollouts.FeatureFlag {
	return t.I(`FeatureFlag`).(*rollouts.FeatureFlag)
}

func GetPilot(t *testcase.T) *rollouts.Pilot {
	return t.I(`Pilot`).(*rollouts.Pilot)
}

func GetPilotEnrollment(t *testcase.T) bool {
	return t.I(`PilotEnrollment`).(bool)
}

func GetRolloutPercentage(t *testcase.T) int {
	return t.I(`RolloutPercentage`).(int)
}

func GetRolloutSeedSalt(t *testcase.T) int64 {
	return t.I(`RolloutSeedSalt`).(int64)
}