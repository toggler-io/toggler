package interactors_test

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRolloutTrier(t *testing.T) {
	t.Parallel()

	var nextRandIntn int
	storage := NewTestStorage()

	trier := func() *interactors.RolloutTrier {
		return &interactors.RolloutTrier{
			Storage: storage,
			RandIntn: func(max int) int {
				return nextRandIntn
			},
		}
	}

	t.Run(`TryRolloutThisPilot`, func(t *testing.T) {
		subject := func() error {
			return trier().TryRolloutThisPilot(flagName, PublicIDOfThePilot)
		}

		cleanAhead := func(t *testing.T) {
			require.Nil(t, storage.Truncate(rollouts.Rollout{}))
			require.Nil(t, storage.Truncate(rollouts.FeatureFlagRolloutStrategy{}))
		}

		t.Run(`when rollout policy is currently not set`, func(t *testing.T) {
			cleanAhead(t)

			t.Run(`then the rollout will return with no erro, and modify nothing`, func(t *testing.T) {
				require.Nil(t, subject())
			})
		})

		t.Run(`when rollout policy is to have a percentage of users to be rolled out`, func(t *testing.T) {
			cleanAhead(t)

			t.Skip()
		})
	})
}
