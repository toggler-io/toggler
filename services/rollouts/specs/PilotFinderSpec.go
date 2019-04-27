package specs

import (
	"strconv"
	"testing"

	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	//nolint:golint,stylecheck
	. "github.com/adamluzsi/FeatureFlags/services/rollouts/testing"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/stretchr/testify/require"
)

type PilotFinderSpec struct {
	Subject interface {
		rollouts.PilotFinder

		specs.MinimumRequirements
	}
}

func (spec PilotFinderSpec) Test(t *testing.T) {
	flagName := ExampleFlagName()

	setup := func(t *testing.T) {
		require.Nil(t, spec.Subject.Truncate(rollouts.FeatureFlag{}))
		require.Nil(t, spec.Subject.Truncate(rollouts.Pilot{}))
	}

	t.Run(`PilotFinderSpec`, func(t *testing.T) {
		var ff *rollouts.FeatureFlag

		subject := func() frameless.Iterator {
			return spec.Subject.FindPilotsByFeatureFlag(ff)
		}

		noPilotsFound := func(t *testing.T) {
			t.Run(`no pilots found`, func(t *testing.T) {
				iter := subject()
				require.NotNil(t, iter)
				require.False(t, iter.Next())
				require.Nil(t, iter.Err())
			})
		}

		t.Run(`when feature was never enabled before`, func(t *testing.T) {
			setup(t)

			ff = nil

			t.Run(`then it will tell that`, noPilotsFound)
		})

		t.Run(`when feature flag exists`, func(t *testing.T) {

			t.Run(`and the flag is not enabled globally`, func(t *testing.T) {
				setup(t)
				ff = &rollouts.FeatureFlag{Name: flagName}
				ff.Rollout.Strategy.Percentage = 0
				require.Nil(t, spec.Subject.Save(ff))

				t.Run(`then it will tell that`, noPilotsFound)

				t.Run(`and the given there is a registered pilot for the feature`, func(t *testing.T) {
					var expectedPilots []*rollouts.Pilot

					for i := 0; i < 5; i++ {
						pilot := &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: strconv.Itoa(i)}
						require.Nil(t, spec.Subject.Save(pilot))
						expectedPilots = append(expectedPilots, pilot)
					}

					t.Run(`then it will return all of them`, func(t *testing.T) {
						iter := subject()
						defer iter.Close()
						require.NotNil(t, iter)

						var actualPilots []*rollouts.Pilot

						for iter.Next() {
							var actually *rollouts.Pilot
							require.Nil(t, iter.Decode(actually))
							actualPilots = append(actualPilots, actually)
						}

						require.Nil(t, iter.Err())

						require.True(t, len(expectedPilots) == len(actualPilots))

						for _, expected := range expectedPilots {
							require.Contains(t, actualPilots, expected)
						}
					})
				})
			})
		})
	})

	t.Run(`FindFlagPilotByExternalPilotID`, func(t *testing.T) {
		var featureFlagID string
		const ExternalPublicPilotID = `42`

		subject := func() (*rollouts.Pilot, error) {
			return spec.Subject.FindFlagPilotByExternalPilotID(featureFlagID, ExternalPublicPilotID)
		}

		noPilotsFound := func(t *testing.T) {
			t.Run(`no pilots found`, func(t *testing.T) {
				pilot, err := subject()
				require.Nil(t, err)
				require.Nil(t, pilot)
			})
		}

		t.Run(`when feature was never enabled before`, func(t *testing.T) {
			require.Nil(t, spec.Subject.Truncate(rollouts.FeatureFlag{}))
			featureFlagID = "not exinsting ID"

			t.Run(`then it will tell that`, noPilotsFound)
		})

		t.Run(`when feature flag exists`, func(t *testing.T) {
			t.Run(`and the flag is not enabled globally`, func(t *testing.T) {
				setup(t)

				ff := &rollouts.FeatureFlag{Name: flagName}
				ff.Rollout.Strategy.Percentage = 100
				require.Nil(t, spec.Subject.Save(ff))
				featureFlagID = ff.ID

				t.Run(`then it will tell that`, noPilotsFound)

				t.Run(`and the given there is a registered pilot for the feature`, func(t *testing.T) {
					require.Nil(t, spec.Subject.Truncate(rollouts.Pilot{}))

					pilot := &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPublicPilotID}
					require.Nil(t, spec.Subject.Save(pilot))

					pilot, err := subject()
					require.Nil(t, err)
					require.NotNil(t, pilot)

					require.Equal(t, ExternalPublicPilotID, pilot.ExternalID)
					require.Equal(t, ff.ID, pilot.FeatureFlagID)
				})
			})
		})
	})
}
