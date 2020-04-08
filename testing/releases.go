package testing

import (
	"context"
	"math/rand"
	"net/url"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
)

const (
	UniqueUserIDLetVar = `UniqUserID`

	ExampleReleaseFlagLetVar     = `ReleaseFlag`
	ExampleReleaseFlagNameLetVar = `ReleaseFlagName`

	ExamplePilotExternalIDLetVar = `PilotExternalID`
	ExamplePilotLetVar           = `Pilot`
	ExamplePilotEnrollmentLetVar = `PilotEnrollment`

	ExampleRolloutPercentageLetVar = `RolloutPercentage`
	ExampleRolloutSeedSaltLetVar   = `RolloutSeedSalt`
	ExampleRolloutApiURLLetVar     = `RolloutApiURL`
)

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(ExampleReleaseFlagNameLetVar, func(t *testcase.T) interface{} {
			return ExampleName()
		})

		s.Let(ExamplePilotExternalIDLetVar, func(t *testcase.T) interface{} {
			return ExampleExternalPilotID()
		})

		s.Let(UniqueUserIDLetVar, func(t *testcase.T) interface{} {
			return ExampleUniqUserID()
		})

		s.Let(ExamplePilotEnrollmentLetVar, func(t *testcase.T) interface{} {
			return rand.Intn(2) == 0
		})

		s.Let(ExamplePilotLetVar, func(t *testcase.T) interface{} {
			return &release.Pilot{
				FlagID:     t.I(ExampleReleaseFlagLetVar).(*release.Flag).ID,
				ExternalID: t.I(ExamplePilotExternalIDLetVar).(string),
				Enrolled:   t.I(ExamplePilotEnrollmentLetVar).(bool),
			}
		})

		s.Let(ExampleRolloutSeedSaltLetVar, func(t *testcase.T) interface{} { return time.Now().Unix() })
		s.Let(ExampleRolloutPercentageLetVar, func(t *testcase.T) interface{} { return int(0) })
		s.Let(ExampleRolloutApiURLLetVar, func(t *testcase.T) interface{} { return nil })

		s.Let(ExampleReleaseFlagLetVar, func(t *testcase.T) interface{} {
			ff := &release.Flag{Name: t.I(ExampleReleaseFlagNameLetVar).(string)}
			ff.Rollout.RandSeed = t.I(`RolloutSeedSalt`).(int64)
			ff.Rollout.Strategy.Percentage = t.I(ExampleRolloutPercentageLetVar).(int)
			ff.Rollout.Strategy.DecisionLogicAPI = GetRolloutApiURL(t)
			return ff
		})
	})
}

// TODO
func GetExternalPilotID(t *testcase.T) string {
	return t.I(ExamplePilotExternalIDLetVar).(string)
}

func ExampleReleaseFlagName(t *testcase.T) string {
	return t.I(ExampleReleaseFlagNameLetVar).(string)
}

func GetPilot(t *testcase.T, vn string) *release.Pilot {
	return t.I(vn).(*release.Pilot)
}

func ExamplePilot(t *testcase.T) *release.Pilot {
	return GetPilot(t, ExamplePilotLetVar)
}

func GetPilotEnrollment(t *testcase.T) bool {
	return t.I(ExamplePilotEnrollmentLetVar).(bool)
}

func GetRolloutPercentage(t *testcase.T) int {
	return t.I(ExampleRolloutPercentageLetVar).(int)
}

func GetRolloutSeedSalt(t *testcase.T) int64 {
	return t.I(ExampleRolloutSeedSaltLetVar).(int64)
}

func GetRolloutApiURL(t *testcase.T) *url.URL {
	rurl := t.I(ExampleRolloutApiURLLetVar)

	if rurl == nil {
		return nil
	}

	u, err := url.Parse(rurl.(string))
	require.Nil(t, err)
	return u
}

func FindStoredExampleReleaseFlagByName(t *testcase.T) *release.Flag {
	return FindStoredReleaseFlagByName(t, ExampleReleaseFlagName(t))
}

func FindStoredReleaseFlagByName(t *testcase.T, name string) *release.Flag {
	f, err := ExampleStorage(t).FindReleaseFlagByName(GetContext(t), name)
	require.Nil(t, err)
	require.NotNil(t, f)
	return f
}

func EnsureFlag(t *testcase.T, name string, prc int) {
	rm := release.NewRolloutManager(ExampleStorage(t))
	require.Nil(t, rm.CreateFeatureFlag(GetContext(t), &release.Flag{
		Name: name,
		Rollout: release.FlagRollout{
			Strategy: release.FlagRolloutStrategy{
				Percentage: prc,
			},
		},
	}))
}

func GetReleaseFlag(t *testcase.T, varName string) *release.Flag {
	ff := t.I(varName)
	if ff == nil {
		return nil
	}
	return ff.(*release.Flag)
}

func ExampleReleaseFlag(t *testcase.T) *release.Flag {
	return GetReleaseFlag(t, ExampleReleaseFlagLetVar)
}

func GetUniqueUserID(t *testcase.T) string {
	return t.I(UniqueUserIDLetVar).(string)
}

func SpecPilotEnrolmentIs(t *testcase.T, enrollment bool) {
	if ExampleReleaseFlag(t).ID == `` {
		require.Nil(t, ExampleStorage(t).Create(context.TODO(), ExampleReleaseFlag(t)))
	}

	rm := release.NewRolloutManager(ExampleStorage(t))
	require.Nil(t, rm.SetPilotEnrollmentForFeature(context.TODO(), ExampleReleaseFlag(t).ID, GetExternalPilotID(t), enrollment, ))
}
