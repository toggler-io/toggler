package testing

import (
	"context"
	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/url"
	"time"
)

func SetupSpecCommonVariables(s *testcase.Spec) {

	s.Let(`ReleaseFlagName`, func(t *testcase.T) interface{} {
		return ExampleName()
	})

	s.Let(`PilotExternalID`, func(t *testcase.T) interface{} {
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
		return &release.Pilot{
			FlagID:     t.I(`ReleaseFlag`).(*release.Flag).ID,
			ExternalID: t.I(`PilotExternalID`).(string),
			Enrolled:   t.I(`PilotEnrollment`).(bool),
		}
	})

	s.Let(`RolloutSeedSalt`, func(t *testcase.T) interface{} { return time.Now().Unix() })
	s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return int(0) })
	s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} { return nil })

	s.Let(`ReleaseFlag`, func(t *testcase.T) interface{} {
		ff := &release.Flag{Name: t.I(`ReleaseFlagName`).(string)}
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
	return t.I(`PilotExternalID`).(string)
}

func GetReleaseFlagName(t *testcase.T) string {
	return t.I(`ReleaseFlagName`).(string)
}

func GetStorage(t *testcase.T) *TestStorage {
	return t.I(`TestStorage`).(*TestStorage)
}

func CTX(t *testcase.T) context.Context {
	return t.I(`ctx`).(context.Context)
}

func GetReleaseFlag(t *testcase.T) *release.Flag {
	ff := t.I(`ReleaseFlag`)

	if ff == nil {
		return nil
	}

	return ff.(*release.Flag)
}

func GetPilot(t *testcase.T) *release.Pilot {
	return t.I(`Pilot`).(*release.Pilot)
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
	if GetReleaseFlag(t).ID == `` {
		require.Nil(t, GetStorage(t).Create(context.TODO(), GetReleaseFlag(t)))
	}

	rm := release.NewRolloutManager(GetStorage(t))
	require.Nil(t, rm.SetPilotEnrollmentForFeature(context.TODO(), GetReleaseFlag(t).ID, GetExternalPilotID(t), enrollment, ))
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

func FindStoredReleaseFlagByName(t *testcase.T) *release.Flag {
	f, err := GetStorage(t).FindReleaseFlagByName(context.TODO(), GetReleaseFlagName(t))
	require.Nil(t, err)
	require.NotNil(t, f)
	return f
}

func EnsureFlag(t *testcase.T, name string, prc int) {
	ctx := context.Background()
	rm := release.NewRolloutManager(GetStorage(t))
	require.Nil(t, rm.CreateFeatureFlag(ctx, &release.Flag{
		Name: name,
		Rollout: release.FlagRollout{
			Strategy: release.FlagRolloutStrategy{
				Percentage: prc,
			},
		},
	}))
}
