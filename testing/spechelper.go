package testing

import (
	"context"
	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/services/rollouts"
	"github.com/toggler-io/toggler/services/security"
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
		ff.Rollout.RandSeed = t.I(`RolloutSeedSalt`).(int64)
		ff.Rollout.Strategy.Percentage = t.I(`RolloutPercentage`).(int)
		ff.Rollout.Strategy.DecisionLogicAPI = GetRolloutApiURL(t)
		return ff
	})

	s.Let(`ctx`, func(t *testcase.T) interface{} {
		return context.Background()
	})

	SetupSpecTokenVariables(s)

}

func SetupSpecTokenVariables(s *testcase.Spec) {

	s.Let(`Token`, func(t *testcase.T) interface{} {
		textToken, objectToken := CreateToken(t, GetUniqUserID(t))
		*(t.I(`*TextToken`).(*string)) = textToken
		return objectToken
	})

	s.Let(`*TextToken`, func(t *testcase.T) interface{} {
		var textToken string
		return &textToken
	})

	s.Let(`TextToken`, func(t *testcase.T) interface{} {
		t.I(`Token`) // trigger *TextTokenSetup
		return *(t.I(`*TextToken`).(*string))
	})

}

func GetTextToken(t *testcase.T) string {
	return t.I(`TextToken`).(string)
}

func GetToken(t *testcase.T) *security.Token {
	return t.I(`Token`).(*security.Token)
}

func CreateToken(t *testcase.T, tokenOwner string) (string, *security.Token) {
	issuer := security.NewIssuer(GetStorage(t))
	textToken, token, err := issuer.CreateNewToken(context.Background(), tokenOwner, nil, nil)
	require.Nil(t, err)
	return textToken, token
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

func CTX(t *testcase.T) context.Context {
	return t.I(`ctx`).(context.Context)
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
	if GetFeatureFlag(t).ID == `` {
		require.Nil(t, GetStorage(t).Save(context.TODO(), GetFeatureFlag(t)))
	}

	rm := rollouts.NewRolloutManager(GetStorage(t))
	require.Nil(t, rm.SetPilotEnrollmentForFeature(context.TODO(), GetFeatureFlag(t).ID, GetExternalPilotID(t), enrollment, ))
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
	f, err := GetStorage(t).FindFlagByName(context.TODO(), GetFeatureFlagName(t))
	require.Nil(t, err)
	require.NotNil(t, f)
	return f
}

func EnsureFlag(t *testcase.T, name string, prc int) {
	rm := rollouts.NewRolloutManager(GetStorage(t))
	require.Nil(t, rm.CreateFeatureFlag(context.TODO(), &rollouts.FeatureFlag{
		Name: name,
		Rollout: rollouts.Rollout{
			Strategy: rollouts.RolloutStrategy{
				Percentage: prc,
			},
		},
	}))
}
