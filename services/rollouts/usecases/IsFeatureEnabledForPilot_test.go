package usecases_test

import (
	"testing"

	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/stretchr/testify/require"

	. "github.com/adamluzsi/FeatureFlags/services/rollouts/testing"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/usecases"
)

func TestIsFeatureEnabledForPilotChecker(t *testing.T) {
	t.Parallel()

	flagName := ExampleFlagName()
	ExternalPilotID := ExampleExternalPilotID()

	storage := NewStorage()

	checker := func() *usecases.IsFeatureEnabledForPilotChecker {
		return usecases.NewIsFeatureEnabledForPilotChecker(storage)
	}

	ensureFlag := func(t *testing.T, flag *rollouts.FeatureFlag) func() {
		require.Nil(t, storage.Save(flag))

		return func() {
			require.Nil(t, storage.DeleteByID(flag, flag.ID))

			flag.ID = ""
		}
	}

	setup := func(t *testing.T) {
		require.Nil(t, storage.Truncate(rollouts.Pilot{}))
		require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))
	}

	pilotEnrollment := func(t *testing.T, ff *rollouts.FeatureFlag) bool {
		require.NotNil(t, ff, `to use this, it is expected that the feature flag already exist`)
		pilot, err := storage.FindFlagPilotByExternalPilotID(ff.ID, ExternalPilotID)

		require.Nil(t, err)
		if pilot == nil {
			return false
		}

		return pilot.Enrolled
	}

	t.Run(`IsFeatureEnabledForPilot`, func(t *testing.T) {
		subject := func() (enabled bool, err error) {
			return checker().IsFeatureEnabledForPilot(flagName, ExternalPilotID)
		}

		t.Run(`when received flag allow enrollment for pilots`, func(t *testing.T) {
			setup(t)
			ff := &rollouts.FeatureFlag{Name: flagName, Rollout: rollouts.Rollout{Percentage: 100, GloballyEnabled: false}}
			defer ensureFlag(t, ff)()

			t.Run(`then pilot on call being enrolled`, func(t *testing.T) {
				enabled, err := subject()

				require.Nil(t, err)
				require.True(t, enabled)
				require.True(t, pilotEnrollment(t, ff))
			})
		})

		t.Run(`when received flag disallow enrollment for pilots`, func(t *testing.T) {
			setup(t)
			ff := &rollouts.FeatureFlag{Name: flagName, Rollout: rollouts.Rollout{Percentage: 0, GloballyEnabled: false}}
			defer ensureFlag(t, ff)()

			t.Run(`then pilot on call being enrolled`, func(t *testing.T) {
				enabled, err := subject()

				require.Nil(t, err)
				require.False(t, enabled)
				require.False(t, pilotEnrollment(t, ff))
			})
		})
	})

}
