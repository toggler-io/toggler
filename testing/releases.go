package testing

import (
	"fmt"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
)

const (
	LetVarExampleReleaseManualPilotEnrollment = `ExampleReleaseManualPilotEnrollment`

	LetVarExampleReleaseRollout  = `example release rollout`
	LetVarExampleReleaseFlag     = `example release flag`
	LetVarExamplePilotExternalID = `PilotExternalID`
	LetVarExamplePilot           = `ManualPilot`
	LetVarExamplePilotEnrollment = `PilotEnrollment`
)

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(LetVarExampleReleaseManualPilotEnrollment, func(t *testcase.T) interface{} {
			mpe := Create(release.ManualPilot{}).(*release.ManualPilot)
			mpe.FlagID = ExampleReleaseFlag(t).ID
			mpe.DeploymentEnvironmentID = ExampleDeploymentEnvironment(t).ID
			mpe.ExternalID = ExampleExternalPilotID(t)
			require.Nil(t, ExampleStorage(t).Create(GetContext(t), mpe))
			t.Defer(ExampleStorage(t).DeleteByID, GetContext(t), release.ManualPilot{}, mpe.ID)
			return mpe
		})

		s.Let(LetVarExamplePilotExternalID, func(t *testcase.T) interface{} {
			return fixtures.Random.StringN(100)
		})

		s.Let(LetVarExamplePilotEnrollment, func(t *testcase.T) interface{} {
			return fixtures.Random.Bool()
		})

		s.Let(LetVarExamplePilot, func(t *testcase.T) interface{} {
			// domains/release/specs/FlagFinder.go:53:1: DEPRECATED, clean it up
			return &release.ManualPilot{
				FlagID:                  ExampleReleaseFlag(t).ID,
				DeploymentEnvironmentID: ExampleDeploymentEnvironment(t).ID,
				ExternalID:              t.I(LetVarExamplePilotExternalID).(string),
				IsParticipating:         t.I(LetVarExamplePilotEnrollment).(bool),
			}
		})

		GivenWeHaveReleaseFlag(s, LetVarExampleReleaseFlag)

		GivenWeHaveReleaseRollout(s,
			LetVarExampleReleaseRollout,
			LetVarExampleReleaseFlag,
			LetVarExampleDeploymentEnvironment,
		)
	})
}

func ExampleReleaseManualPilotEnrollment(t *testcase.T) *release.ManualPilot {
	return t.I(LetVarExampleReleaseManualPilotEnrollment).(*release.ManualPilot)
}

func ExampleExternalPilotID(t *testcase.T) string {
	return t.I(LetVarExamplePilotExternalID).(string)
}

func FindStoredReleaseFlagByName(t *testcase.T, name string) *release.Flag {
	f, err := ExampleStorage(t).FindReleaseFlagByName(GetContext(t), name)
	require.Nil(t, err)
	require.NotNil(t, f)
	return f
}

func ExampleReleaseRollout(t *testcase.T) *release.Rollout {
	return GetReleaseRollout(t, LetVarExampleReleaseRollout)
}

func getReleaseRolloutPlanLetVar(vn string) string {
	return fmt.Sprintf(`%s.plan`, vn)
}

func GetReleaseRolloutPlan(t *testcase.T, rolloutLVN string) release.RolloutDefinition {
	return t.I(getReleaseRolloutPlanLetVar(rolloutLVN)).(release.RolloutDefinition)
}

func GetReleaseRollout(t *testcase.T, vn string) *release.Rollout {
	return t.I(vn).(*release.Rollout)
}

func GivenWeHaveReleaseRollout(s *testcase.Spec, vn, flagLVN, envLVN string) {
	s.Let(getReleaseRolloutPlanLetVar(vn), func(t *testcase.T) interface{} {
		return *Create(release.RolloutDecisionByPercentage{}).(*release.RolloutDecisionByPercentage)
	})

	s.Let(vn, func(t *testcase.T) interface{} {
		rf := GetReleaseFlag(t, flagLVN)
		de := GetDeploymentEnvironment(t, envLVN)

		rollout := FixtureFactory{}.Create(release.Rollout{}).(*release.Rollout)
		rollout.FlagID = rf.ID
		rollout.DeploymentEnvironmentID = de.ID
		rollout.Plan = GetReleaseRolloutPlan(t, vn)
		require.Nil(t, rollout.Plan.Validate())

		// TODO: replace when rollout manager has function for this
		require.Nil(t, ExampleRolloutManager(t).Storage.Create(GetContext(t), rollout))
		t.Defer(ExampleRolloutManager(t).Storage.DeleteByID, GetContext(t), *rollout, rollout.ID)
		t.Logf(`%#v`, rollout)
		return rollout
	})
}

func GivenWeHaveReleaseFlag(s *testcase.Spec, vn string) {
	s.Let(vn, func(t *testcase.T) interface{} {
		rf := FixtureFactory{}.Create(release.Flag{}).(*release.Flag)
		rf.Name = fmt.Sprintf(`%s - %s`, vn, rf.Name)
		require.Nil(t, ExampleRolloutManager(t).Storage.Create(GetContext(t), rf))
		t.Defer(ExampleRolloutManager(t).DeleteFeatureFlag, GetContext(t), rf.ID)
		t.Defer(ExampleStorage(t).DeleteByID, GetContext(t), release.Flag{}, rf.ID)
		t.Logf(`%#v`, rf)
		return rf
	})
}

func AndReleaseFlagRolloutPercentageIs(s *testcase.Spec, rolloutLVN string, percentage int) {
	s.Before(func(t *testcase.T) {
		rollout := GetReleaseRollout(t, LetVarExampleReleaseRollout)
		byPercentage, ok := rollout.Plan.(release.RolloutDecisionByPercentage)
		require.True(t, ok, `unexpected release rollout plan definition for AndReleaseFlagRolloutPercentageIs helper`)
		byPercentage.Percentage = percentage
		t.Logf(`and the release rollout percentage is set to %d`, percentage)
		//
		// please note that this will eager load the rollout value in the testing framework
		// as it makes no sense to have a percentage set to something that doesn't even exists.
		//
		// And in case if we already initialized such context where rollout entry exists,
		// we need to update its rollout plan as well.
		t.Let(getReleaseRolloutPlanLetVar(rolloutLVN), byPercentage)
		rollout.Plan = GetReleaseRolloutPlan(t, LetVarExampleReleaseRollout)
		require.Nil(t, ExampleStorage(t).Update(GetContext(t), rollout))
	})
}

func GetReleaseFlag(t *testcase.T, lvn string) *release.Flag {
	ff := t.I(lvn)
	if ff == nil {
		return nil
	}
	return ff.(*release.Flag)
}

func ExampleReleaseFlag(t *testcase.T) *release.Flag {
	return GetReleaseFlag(t, LetVarExampleReleaseFlag)
}

func ExampleRolloutManager(t *testcase.T) *release.RolloutManager {
	return release.NewRolloutManager(ExampleStorage(t))
}

func SpecPilotEnrolmentIs(t *testcase.T, enrollment bool) {
	if ExampleReleaseFlag(t).ID == `` {
		require.Nil(t, ExampleStorage(t).Create(GetContext(t), ExampleReleaseFlag(t)))
	}

	rm := release.NewRolloutManager(ExampleStorage(t))
	require.Nil(t, rm.SetPilotEnrollmentForFeature(GetContext(t),
		ExampleReleaseFlag(t).ID,
		ExampleDeploymentEnvironment(t).ID,
		ExampleExternalPilotID(t),
		enrollment))
}

func NoReleaseFlagPresentInTheStorage(s *testcase.Spec) {
	s.Before(func(t *testcase.T) {
		// TODO: replace with flag manager list+delete
		require.Nil(t, ExampleStorage(t).DeleteAll(GetContext(t), release.Flag{}))
	})
}

func NoReleaseRolloutPresentInTheStorage(s *testcase.Spec) {
	s.Before(func(t *testcase.T) {
		// TODO: replace with rollout manager list+delete
		require.Nil(t, ExampleStorage(t).DeleteAll(GetContext(t), release.Rollout{}))
	})
}

func AndExamplePilotManualParticipatingIsSetTo(s *testcase.Spec, isParticipating bool) {
	s.Before(func(t *testcase.T) {
		require.Nil(t, ExampleRolloutManager(t).SetPilotEnrollmentForFeature(
			GetContext(t),
			ExampleReleaseFlag(t).ID,
			ExampleDeploymentEnvironment(t).ID,
			ExampleExternalPilotID(t),
			isParticipating,
		))

		t.Defer(ExampleRolloutManager(t).UnsetPilotEnrollmentForFeature,
			GetContext(t),
			ExampleReleaseFlag(t).ID,
			ExampleDeploymentEnvironment(t).ID,
			ExampleExternalPilotID(t),
		)
	})
}
