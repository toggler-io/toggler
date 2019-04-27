package interactors_test

import (
	"testing"
	"time"

	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
	. "github.com/adamluzsi/FeatureFlags/services/rollouts/testing"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/stretchr/testify/require"
)

func TestRolloutManager(t *testing.T) {
	t.Parallel()

	var ff *rollouts.FeatureFlag

	ExternalPilotID := ExampleExternalPilotID()
	FeatureFlagName := ExampleFlagName()
	GeneratedRandomSeed := time.Now().Unix()
	storage := NewStorage()

	manager := func() *interactors.RolloutManager {
		return &interactors.RolloutManager{
			Storage: storage,
			RandSeedGenerator: func() int64 {
				return GeneratedRandomSeed
			},
		}
	}

	setup := func(t *testing.T, ffSetup func(*rollouts.FeatureFlag)) {
		require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))
		require.Nil(t, storage.Truncate(rollouts.Pilot{}))

		ff = &rollouts.FeatureFlag{Name: FeatureFlagName}

		require.Nil(t, storage.Save(ff))

		if ffSetup != nil {
			ffSetup(ff)
		}

	}

	t.Run(`EnableFeatureFor`, func(t *testing.T) {
		subject := func() error { return manager().EnableFeatureFor(FeatureFlagName, ExternalPilotID) }

		findFlag := func(t *testing.T) *rollouts.FeatureFlag {
			iter := storage.FindAll(&rollouts.FeatureFlag{})
			require.NotNil(t, iter)
			require.True(t, iter.Next())
			var ff rollouts.FeatureFlag
			require.Nil(t, iter.Decode(&ff))
			require.False(t, iter.Next())
			require.Nil(t, iter.Err())
			return &ff
		}

		t.Run(`when no feature flag is seen ever before`, func(t *testing.T) {

			flagSetup := func(flag *rollouts.FeatureFlag) {
				require.Nil(t, storage.DeleteByID(rollouts.FeatureFlag{}, flag.ID))
			}

			t.Run(`then feature flag created`, func(t *testing.T) {
				setup(t, flagSetup)

				require.Nil(t, subject())

				flag := findFlag(t)

				require.Equal(t, FeatureFlagName, flag.Name)
				require.Equal(t, "", flag.Rollout.Strategy.URL)
				require.Equal(t, 0, flag.Rollout.Strategy.Percentage)
				require.Equal(t, GeneratedRandomSeed, flag.Rollout.RandSeedSalt)
			})

			t.Run(`then pilot is enrolled for the feature`, func(t *testing.T) {
				setup(t, flagSetup)

				require.Nil(t, subject())
				flag := findFlag(t)

				pilot, err := storage.FindFlagPilotByExternalPilotID(flag.ID, ExternalPilotID)
				require.Nil(t, err)

				require.NotNil(t, pilot)
				require.Equal(t, true, pilot.Enrolled)
				require.Equal(t, ExternalPilotID, pilot.ExternalID)

			})

		})

		t.Run(`when feature flag already configured`, func(t *testing.T) {

			flagSetup := func(flag *rollouts.FeatureFlag) {
				require.NotNil(t, flag)
				require.NotEmpty(t, flag.ID)
			}

			t.Run(`then flag is not recreated`, func(t *testing.T) {
				setup(t, flagSetup)

				require.Nil(t, subject())

				count, err := iterators.Count(storage.FindAll(rollouts.FeatureFlag{}))
				require.Nil(t, err)
				require.Equal(t, 1, count)

				flag := findFlag(t)
				require.Equal(t, ff, flag)
			})

			t.Run(`and pilot already exists`, func(t *testing.T) {
				t.Run(`and pilot is already enrolled`, func(t *testing.T) {
					setup(t, flagSetup)
					originalPilot := &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPilotID, Enrolled: true}
					require.Nil(t, storage.Save(originalPilot))

					t.Run(`then original pilot is kept for enrollment`, func(t *testing.T) {

						require.Nil(t, subject())
						flag := findFlag(t)

						pilot, err := storage.FindFlagPilotByExternalPilotID(flag.ID, ExternalPilotID)
						require.Nil(t, err)

						require.NotNil(t, pilot)
						require.Equal(t, true, pilot.Enrolled)
						require.Equal(t, ExternalPilotID, pilot.ExternalID)
						require.Equal(t, originalPilot, pilot)

						count, err := iterators.Count(storage.FindAll(rollouts.Pilot{}))
						require.Nil(t, err)
						require.Equal(t, 1, count)
					})
				})

				t.Run(`and pilot is blacklisted currently`, func(t *testing.T) {
					setup(t, flagSetup)
					require.Nil(t, storage.Save(&rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPilotID, Enrolled: false}))

					t.Run(`then pilot is enrolled`, func(t *testing.T) {

						require.Nil(t, subject())
						ff := findFlag(t)

						pilot, err := storage.FindFlagPilotByExternalPilotID(ff.ID, ExternalPilotID)
						require.Nil(t, err)

						require.NotNil(t, pilot)
						require.Equal(t, true, pilot.Enrolled)
						require.Equal(t, ExternalPilotID, pilot.ExternalID)

						count, err := iterators.Count(storage.FindAll(rollouts.Pilot{}))
						require.Nil(t, err)
						require.Equal(t, 1, count)

					})
				})
			})
		})
	})
}
