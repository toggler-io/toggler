package usecases_test

import (
	"testing"

	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/stretchr/testify/require"

	. "github.com/adamluzsi/FeatureFlags/services/rollouts/testing"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/usecases"
)

func TestIsFeatureGloballyEnabledChecker(t *testing.T) {
	t.Parallel()

	flagName := ExampleFlagName()

	storage := NewStorage()

	checker := func() *usecases.IsFeatureGloballyEnabledChecker {
		return usecases.NewIsFeatureGloballyEnabledChecker(storage)
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

	t.Run(`IsFeatureGloballyEnabled`, func(t *testing.T) {

		subject := func() (enabled bool, err error) {
			return checker().IsFeatureGloballyEnabled(flagName)
		}

		t.Run(`when received flag name is rolled out globally already`, func(t *testing.T) {
			setup(t)
			ff := &rollouts.FeatureFlag{Name: flagName, Rollout: rollouts.Rollout{GloballyEnabled: true}}
			defer ensureFlag(t, ff)()

			t.Run(`then it will be reported as enabled`, func(t *testing.T) {
				enabled, err := subject()

				require.Nil(t, err)
				require.True(t, enabled)
			})
		})

	})

}
