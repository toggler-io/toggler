package interactors_test

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsFeatureEnabled(t *testing.T) {
	t.Parallel()

	storage := NewTestStorage()
	subject := func() (bool, error) {
		return interactors.IsFeatureEnabled(storage, flagName, PublicIDOfThePilot)
	}

	t.Run(`when feature was never enabled before`, func(t *testing.T) {
		require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))

		t.Run(`then it will tell that feature flag is not enabled`, func(t *testing.T) {
			ok, err := subject()
			require.Nil(t, err)
			require.False(t, ok)
		})
	})

	t.Run(`when feature flag exists`, func(t *testing.T) {
		t.Run(`and the flag is not enabled globally`, func(t *testing.T) {
			ff := &rollouts.FeatureFlag{Name: flagName, IsGlobal: false}
			require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))
			require.Nil(t, storage.Save(ff))

			t.Run(`then it will tell that feature flag is not enabled`, func(t *testing.T) {
				ok, err := subject()
				require.Nil(t, err)
				require.False(t, ok)
			})

			t.Run(`and the given user is enabled for piloting the feature`, func(t *testing.T) {
				pilot := &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalPublicID: PublicIDOfThePilot}
				require.Nil(t, storage.Save(pilot))

				t.Run(`then it will tell that feature flag is enabled`, func(t *testing.T) {
					ok, err := subject()
					require.Nil(t, err)
					require.True(t, ok)
				})
			})
		})

		t.Run(`and the flag is enabled globally`, func(t *testing.T) {
			ff := &rollouts.FeatureFlag{Name: flagName, IsGlobal: true}
			require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))
			require.Nil(t, storage.Save(ff))

			t.Run(`then it will tell that feature flag is enabled`, func(t *testing.T) {
				ok, err := subject()
				require.Nil(t, err)
				require.True(t, ok)
			})

			t.Run(`and the given user is enabled for piloting the feature`, func(t *testing.T) {
				pilot := &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalPublicID: PublicIDOfThePilot}
				require.Nil(t, storage.Save(pilot))

				t.Run(`then it will tell that feature flag is enabled`, func(t *testing.T) {
					ok, err := subject()
					require.Nil(t, err)
					require.True(t, ok)
				})
			})
		})
	})
}
