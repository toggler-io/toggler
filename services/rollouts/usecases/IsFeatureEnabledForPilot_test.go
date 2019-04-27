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

	t.Run(`IsFeatureEnabledForPilot`, func(t *testing.T) {
		subject := func() (enabled bool, err error) {
			return checker().IsFeatureEnabledForPilot(flagName, ExternalPilotID)
		}

		t.Run(`when received flag allow enrollment for pilots`, func(t *testing.T) {
			setup(t)
			ff := &rollouts.FeatureFlag{Name: flagName, Rollout: rollouts.Rollout{Percentage: 100}}
			defer ensureFlag(t, ff)()

			t.Run(`then pilot on call being enrolled`, func(t *testing.T) {
				enabled, err := subject()

				require.Nil(t, err)
				require.True(t, enabled)
			})
		})

		t.Run(`when received flag disallow enrollment for pilots`, func(t *testing.T) {
			setup(t)
			ff := &rollouts.FeatureFlag{Name: flagName, Rollout: rollouts.Rollout{Percentage: 0}}
			defer ensureFlag(t, ff)()

			t.Run(`then pilot on call being enrolled`, func(t *testing.T) {
				enabled, err := subject()

				require.Nil(t, err)
				require.False(t, enabled)
			})
		})
	})

}
