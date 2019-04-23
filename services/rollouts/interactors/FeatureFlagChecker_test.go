package interactors_test

import (
	"testing"

	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
	. "github.com/adamluzsi/FeatureFlags/services/rollouts/testing"
	"github.com/stretchr/testify/require"
)

func TestFeatureFlagChecker(t *testing.T) {
	t.Parallel()

	ExternalPilotID := ExampleExternalPilotID()
	FlagName := ExampleFlagName()

	storage := NewStorage()

	featureFlagChecker := func() *interactors.FeatureFlagChecker {
		return &interactors.FeatureFlagChecker{Storage: storage}
	}

	setup := func(t *testing.T) {
		require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))
		require.Nil(t, storage.Truncate(rollouts.Pilot{}))
	}

	pilotIs := func(t *testing.T, ff *rollouts.FeatureFlag, pilot *rollouts.Pilot) func() {
		require.Nil(t, storage.Save(pilot))
		return func() { require.Nil(t, storage.DeleteByID(rollouts.Pilot{}, pilot.ID)) }
	}

	t.Run(`IsFeatureEnabledFor`, func(t *testing.T) {
		subject := func() (bool, error) {
			return featureFlagChecker().IsFeatureEnabledFor(FlagName, ExternalPilotID)
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

				ff := &rollouts.FeatureFlag{Name: FlagName}
				ff.Rollout.GloballyEnabled = false
				require.Nil(t, storage.Save(ff))

				t.Run(`then it will tell that feature flag is not enabled`, func(t *testing.T) {
					enabled, err := subject()
					require.Nil(t, err)
					require.False(t, enabled)
				})

				t.Run(`and the given user is enabled for piloting the feature`, func(t *testing.T) {
					defer pilotIs(t, ff, &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPilotID, Enrolled: true})()

					t.Run(`then it will tell that feature flag is enabled`, func(t *testing.T) {
						enabled, err := subject()
						require.Nil(t, err)
						require.True(t, enabled)
					})
				})

				t.Run(`and the given user was disabled from piloting the feature`, func(t *testing.T) {
					defer pilotIs(t, ff, &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPilotID, Enrolled: false})()

					t.Run(`then it will tell that feature flag is disabled`, func(t *testing.T) {
						enabled, err := subject()
						require.Nil(t, err)
						require.False(t, enabled)
					})
				})
			})

			t.Run(`and the flag is enabled globally`, func(t *testing.T) {
				setup(t)

				ff := &rollouts.FeatureFlag{Name: FlagName}
				ff.Rollout.GloballyEnabled = true
				require.Nil(t, storage.Save(ff))

				t.Run(`then it will tell that feature flag is enabled`, func(t *testing.T) {
					enabled, err := subject()
					require.Nil(t, err)
					require.True(t, enabled)
				})

				t.Run(`and the given user is enabled for piloting the feature`, func(t *testing.T) {
					defer pilotIs(t, ff, &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPilotID, Enrolled: true})()

					t.Run(`then it will tell that feature flag is enabled`, func(t *testing.T) {
						enabled, err := subject()
						require.Nil(t, err)
						require.True(t, enabled)
					})
				})

				t.Run(`and the given user was disabled from piloting the feature`, func(t *testing.T) {
					defer pilotIs(t, ff, &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPilotID, Enrolled: false})()

					t.Run(`then it will tell that feature flag is enabled`, func(t *testing.T) {
						t.Log(`this is because Pilot Enroll false is not a blacklist`)
						enabled, err := subject()
						require.Nil(t, err)
						require.True(t, enabled)
					})
				})
			})
		})
	})

	t.Run(`IsFeatureGloballyEnabled`, func(t *testing.T) {
		subject := func() (bool, error) {
			return featureFlagChecker().IsFeatureGloballyEnabled(FlagName)
		}

		thenItWillReportThatFeatureNotGlobballyEnabled := func(t *testing.T) {
			t.Run(`then it will report that feature is not enabled globally`, func(t *testing.T) {
				enabled, err := subject()
				require.Nil(t, err)
				require.False(t, enabled)
			})
		}

		thenItWillReportThatFeatureIsGloballyRolledOut := func(t *testing.T) {
			t.Run(`then it will report that feature is globally rolled out`, func(t *testing.T) {
				enabled, err := subject()
				require.Nil(t, err)
				require.True(t, enabled)
			})
		}

		t.Run(`when feature flag is not seen before`, func(t *testing.T) {
			setup(t)

			thenItWillReportThatFeatureNotGlobballyEnabled(t)
		})

		t.Run(`when feature flag is given`, func(t *testing.T) {
			t.Run(`and it is not yet rolled out globally`, func(t *testing.T) {
				setup(t)

				ff := &rollouts.FeatureFlag{Name: FlagName}
				ff.Rollout.GloballyEnabled = false
				require.Nil(t, storage.Save(ff))

				thenItWillReportThatFeatureNotGlobballyEnabled(t)
			})

			t.Run(`and it is rolled out globally`, func(t *testing.T) {
				setup(t)

				ff := &rollouts.FeatureFlag{Name: FlagName}
				ff.Rollout.GloballyEnabled = true
				require.Nil(t, storage.Save(ff))

				thenItWillReportThatFeatureIsGloballyRolledOut(t)
			})
		})

	})
}
