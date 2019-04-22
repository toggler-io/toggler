package interactors_test

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFeatureFlagChecker(t *testing.T) {
	t.Parallel()

	PublicIDOfThePilot := randomdata.MacAddress()
	flagName := randomdata.SillyName()

	storage := NewTestStorage()

	featureFlagChecker := func() *interactors.FeatureFlagChecker {
		return &interactors.FeatureFlagChecker{Storage: storage}
	}

	setup := func(t *testing.T) {
		require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))
		require.Nil(t, storage.Truncate(rollouts.Pilot{}))
	}

	t.Run(`IsFeatureEnabled`, func(t *testing.T) {
		subject := func() (bool, error) {
			return featureFlagChecker().IsFeatureEnabled(flagName, PublicIDOfThePilot)
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
				setup(t)

				ff := &rollouts.FeatureFlag{Name: flagName}
				ff.Rollout.GloballyEnabled = false
				require.Nil(t, storage.Save(ff))

				t.Run(`then it will tell that feature flag is not enabled`, func(t *testing.T) {
					enabled, err := subject()
					require.Nil(t, err)
					require.False(t, enabled)
				})

				t.Run(`and the given user is enabled for piloting the feature`, func(t *testing.T) {
					pilot := &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalPublicID: PublicIDOfThePilot}
					require.Nil(t, storage.Save(pilot))

					t.Run(`then it will tell that feature flag is enabled`, func(t *testing.T) {
						enabled, err := subject()
						require.Nil(t, err)
						require.True(t, enabled)
					})
				})
			})

			t.Run(`and the flag is enabled globally`, func(t *testing.T) {
				setup(t)

				ff := &rollouts.FeatureFlag{Name: flagName}
				ff.Rollout.GloballyEnabled = true
				require.Nil(t, storage.Save(ff))

				t.Run(`then it will tell that feature flag is enabled`, func(t *testing.T) {
					enabled, err := subject()
					require.Nil(t, err)
					require.True(t, enabled)
				})

				t.Run(`and the given user is enabled for piloting the feature`, func(t *testing.T) {
					pilot := &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalPublicID: PublicIDOfThePilot}
					require.Nil(t, storage.Save(pilot))

					t.Run(`then it will tell that feature flag is enabled`, func(t *testing.T) {
						enabled, err := subject()
						require.Nil(t, err)
						require.True(t, enabled)
					})
				})
			})
		})
	})
}
