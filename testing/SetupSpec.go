package testing

import (
	"math/rand"
	"time"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/testcase"
)

func SetupSpecCommonVariables(s *testcase.Spec) {

	s.Let(`FeatureName`, func(v *testcase.V) interface{} {
		return ExampleFeatureName()
	})

	s.Let(`ExternalPilotID`, func(v *testcase.V) interface{} {
		return ExampleExternalPilotID()
	})

	s.Let(`Storage`, func(v *testcase.V) interface{} {
		return NewStorage()
	})

	s.Let(`PilotEnrollment`, func(v *testcase.V) interface{} {
		return rand.Intn(2) == 0
	})

	s.Let(`Pilot`, func(v *testcase.V) interface{} {
		return &rollouts.Pilot{
			FeatureFlagID: v.I(`FeatureFlag`).(*rollouts.FeatureFlag).ID,
			ExternalID:    v.I(`ExternalPilotID`).(string),
			Enrolled:      v.I(`PilotEnrollment`).(bool),
		}
	})

	s.Let(`RolloutSeedSalt`, func(v *testcase.V) interface{} { return time.Now().Unix() })
	s.Let(`RolloutPercentage`, func(v *testcase.V) interface{} { return int(0) })
	s.Let(`RolloutApiURL`, func(v *testcase.V) interface{} { return "" })
	s.Let(`FeatureFlag`, func(v *testcase.V) interface{} {
		ff := &rollouts.FeatureFlag{Name: v.I(`FeatureName`).(string)}
		ff.Rollout.RandSeedSalt = v.I(`RolloutSeedSalt`).(int64)
		ff.Rollout.Strategy.Percentage = v.I(`RolloutPercentage`).(int)
		ff.Rollout.Strategy.URL = v.I(`RolloutApiURL`).(string)
		return ff
	})

}

func GetExternalPilotID(v *testcase.V) string {
	return v.I(`ExternalPilotID`).(string)
}

func GetFeatureFlagName(v *testcase.V) string {
	return v.I(`FeatureName`).(string)
}

func GetStorage(v *testcase.V) *Storage {
	return v.I(`Storage`).(*Storage)
}

func GetFeatureFlag(v *testcase.V) *rollouts.FeatureFlag {
	return v.I(`FeatureFlag`).(*rollouts.FeatureFlag)
}

func GetPilot(v *testcase.V) *rollouts.Pilot {
	return v.I(`Pilot`).(*rollouts.Pilot)
}

func GetPilotEnrollment(v *testcase.V) bool {
	return v.I(`PilotEnrollment`).(bool)
}

func GetRolloutPercentage(v *testcase.V) int {
	return v.I(`RolloutPercentage`).(int)
}

func GetRolloutSeedSalt(v *testcase.V) int64 {
	return v.I(`RolloutSeedSalt`).(int64)
}