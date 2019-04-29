package rollouts_test

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"math/rand"
	"testing"
	"time"

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

	manager := func() *rollouts.RolloutManager {
		return &rollouts.RolloutManager{
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

	t.Run(`SetPilotEnrollmentForFeature`, func(t *testing.T) {
		enrollment := rand.Intn(2) == 0

		subject := func() error { return manager().SetPilotEnrollmentForFeature(FeatureFlagName, ExternalPilotID, enrollment) }

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

			t.Run(`then pilot is enrollment for the feature is set`, func(t *testing.T) {
				setup(t, flagSetup)

				require.Nil(t, subject())
				flag := findFlag(t)

				pilot, err := storage.FindFlagPilotByExternalPilotID(flag.ID, ExternalPilotID)
				require.Nil(t, err)

				require.NotNil(t, pilot)
				require.Equal(t, enrollment, pilot.Enrolled)
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
				t.Run(`and pilot is has the opposite enrollment status`, func(t *testing.T) {
					setup(t, flagSetup)
					originalPilot := &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPilotID, Enrolled: !enrollment}
					require.Nil(t, storage.Save(originalPilot))

					t.Run(`then original pilot is updated to the new enrollment status`, func(t *testing.T) {

						require.Nil(t, subject())
						flag := findFlag(t)

						pilot, err := storage.FindFlagPilotByExternalPilotID(flag.ID, ExternalPilotID)
						require.Nil(t, err)

						require.NotNil(t, pilot)
						require.Equal(t, enrollment, pilot.Enrolled)
						require.Equal(t, ExternalPilotID, pilot.ExternalID)
						require.Equal(t, originalPilot, pilot)

						count, err := iterators.Count(storage.FindAll(rollouts.Pilot{}))
						require.Nil(t, err)
						require.Equal(t, 1, count)
					})
				})

				t.Run(`and pilot already has the same enrollment status`, func(t *testing.T) {
					setup(t, flagSetup)
					require.Nil(t, storage.Save(&rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPilotID, Enrolled: enrollment}))

					t.Run(`then pilot remain the same`, func(t *testing.T) {

						require.Nil(t, subject())
						ff := findFlag(t)

						pilot, err := storage.FindFlagPilotByExternalPilotID(ff.ID, ExternalPilotID)
						require.Nil(t, err)

						require.NotNil(t, pilot)
						require.Equal(t, enrollment, pilot.Enrolled)
						require.Equal(t, ExternalPilotID, pilot.ExternalID)

						count, err := iterators.Count(storage.FindAll(rollouts.Pilot{}))
						require.Nil(t, err)
						require.Equal(t, 1, count)

					})
				})
			})
		})
	})

	t.Run(`UpdateFeatureFlagRolloutPercentage`, func(t *testing.T) {
		var RolloutPercentage int
		subject := func() error {
			return manager().UpdateFeatureFlagRolloutPercentage(FeatureFlagName, RolloutPercentage)
		}

		t.Run(`when percentage less than 0`, func(t *testing.T) {
			RolloutPercentage = -1 + (rand.Intn(1024) * -1)

			t.Run(`then it will report error`, func(t *testing.T) {
				require.Error(t, subject())
			})
		})
		t.Run(`when percentage greater than 100`, func(t *testing.T) {
			RolloutPercentage = 101 + rand.Intn(1024)

			t.Run(`then it will report error`, func(t *testing.T) {
				require.Error(t, subject())
			})
		})
		t.Run(`when percentage is a number between 0 and 100`, func(t *testing.T) {
			RolloutPercentage = rand.Intn(101)
			getRandomPercentageThatIsNotEqualWith := func(oth int) int {
				for {
					roll := rand.Intn(101)
					if roll != RolloutPercentage {
						return roll
					}
				}
			}

			t.Run(`and feature flag was undefined until now`, func(t *testing.T) {
				ffSetup := func(flag *rollouts.FeatureFlag) {
					require.Nil(t, storage.DeleteByID(flag, flag.ID))
				}

				t.Run(`then feature flag entry created with the percentage`, func(t *testing.T) {
					setup(t, ffSetup)

					require.Nil(t, subject())
					flag, err := storage.FindByFlagName(FeatureFlagName)
					require.Nil(t, err)
					require.NotNil(t, flag)
					require.Equal(t, FeatureFlagName, flag.Name)
					require.Equal(t, "", flag.Rollout.Strategy.URL)
					require.Equal(t, RolloutPercentage, flag.Rollout.Strategy.Percentage)
					require.Equal(t, GeneratedRandomSeed, flag.Rollout.RandSeedSalt)
				})
			})
			t.Run(`and feature flag is already exist with a different percentage`, func(t *testing.T) {
				originalPercentage := getRandomPercentageThatIsNotEqualWith(RolloutPercentage)
				originalSeedSalt := GeneratedRandomSeed + 1
				ffSetup := func(flag *rollouts.FeatureFlag) {
					ff.Rollout.Strategy.Percentage = originalPercentage
					ff.Rollout.RandSeedSalt = originalSeedSalt
					require.Nil(t, storage.Update(ff))
				}

				t.Run(`then the same feature flag kept but updated to the new percentage`, func(t *testing.T) {
					setup(t, ffSetup)

					require.Nil(t, subject())
					flag, err := storage.FindByFlagName(FeatureFlagName)
					require.Nil(t, err)
					require.NotNil(t, flag)
					require.Equal(t, ff.ID, flag.ID)
					require.Equal(t, FeatureFlagName, flag.Name)
					require.Equal(t, "", flag.Rollout.Strategy.URL)
					require.Equal(t, RolloutPercentage, flag.Rollout.Strategy.Percentage)
					require.Equal(t, originalSeedSalt, flag.Rollout.RandSeedSalt)
				})
			})
		})
	})
}
