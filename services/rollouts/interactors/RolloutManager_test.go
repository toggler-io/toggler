package interactors_test

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRolloutTrier(t *testing.T) {
	t.Parallel()

	PublicIDOfThePilot := randomdata.MacAddress()
	flagName := randomdata.SillyName()

	var nextRandIntn int
	storage := NewTestStorage()

	trier := func() *interactors.RolloutManager {
		return &interactors.RolloutManager{
			Storage: storage,
			RandIntn: func(max int) int {
				return nextRandIntn
			},
		}
	}

	setup := func(t *testing.T) {
		require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))
		require.Nil(t, storage.Truncate(rollouts.Pilot{}))
	}

	t.Run(`TryRolloutThisPilot`, func(t *testing.T) {
		subject := func() error {
			return trier().TryRolloutThisPilot(flagName, PublicIDOfThePilot)
		}

		t.Run(`when rollout policy is currently not set`, func(t *testing.T) {
			setup(t)

			t.Run(`then the rollout will return with no erro, and modify nothing`, func(t *testing.T) {
				require.Nil(t, subject())
			})
		})

		t.Run(`when rollout policy is to have a percentage of users to be rolled out`, func(t *testing.T) {
			setup(t)

			t.Skip()
		})
	})
}
