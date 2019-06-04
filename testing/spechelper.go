package testing

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/security"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/url"
	"time"
)

func SetupSpecCommonVariables(s *testcase.Spec) {

	s.Let(`FeatureName`, func(t *testcase.T) interface{} {
		return ExampleFeatureName()
	})

	s.Let(`ExternalPilotID`, func(t *testcase.T) interface{} {
		return ExampleExternalPilotID()
	})

	s.Let(`UniqUserID`, func(t *testcase.T) interface{} {
		return ExampleUniqUserID()
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
	s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} { return nil })

	s.Let(`FeatureFlag`, func(t *testcase.T) interface{} {
		ff := &rollouts.FeatureFlag{Name: t.I(`FeatureName`).(string)}
		ff.Rollout.RandSeedSalt = t.I(`RolloutSeedSalt`).(int64)
		ff.Rollout.Strategy.Percentage = t.I(`RolloutPercentage`).(int)
		ff.Rollout.Strategy.DecisionLogicAPI = GetRolloutApiURL(t)
		return ff
	})

}

func CreateToken(t *testcase.T, tokenOwner string) *security.Token {
	i := security.NewIssuer(GetStorage(t))
	token, err := i.CreateNewToken(tokenOwner, nil, nil)
	require.Nil(t, err)
	return token
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
	ff := t.I(`FeatureFlag`)

	if ff == nil {
		return nil
	}

	return ff.(*rollouts.FeatureFlag)
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

func GetUniqUserID(t *testcase.T) string {
	return t.I(`UniqUserID`).(string)
}

func SpecPilotEnrolmentIs(t *testcase.T, enrollment bool) {
	rm := rollouts.NewRolloutManager(GetStorage(t))
	require.Nil(t, rm.SetPilotEnrollmentForFeature(
		GetFeatureFlagName(t),
		GetExternalPilotID(t),
		enrollment,
	))
}

func GetRolloutApiURL(t *testcase.T) *url.URL {
	rurl := t.I(`RolloutApiURL`)

	if rurl == nil {
		return nil
	}

	u, err := url.Parse(rurl.(string))
	require.Nil(t, err)
	return u
}

func FindStoredFeatureFlagByName(t *testcase.T) *rollouts.FeatureFlag {
	f, err := GetStorage(t).FindFlagByName(GetFeatureFlagName(t))
	require.Nil(t, err)
	require.NotNil(t, f)
	return f
}

func EnsureFlag(t *testcase.T, name string, prc int) {
	rm := rollouts.NewRolloutManager(GetStorage(t))
	require.Nil(t, rm.CreateFeatureFlag(&rollouts.FeatureFlag{
		Name: name,
		Rollout: rollouts.Rollout{
			Strategy: rollouts.RolloutStrategy{
				Percentage: prc,
			},
		},
	}))
}
