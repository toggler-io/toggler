package rollouts_test

import (
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/testcase"
	"math/rand"
	"testing"
	"time"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/stretchr/testify/require"
)

func TestRolloutManager(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()
	SetupSpecCommonVariables(s)

	GetGeneratedRandomSeed := func(v *testcase.V) int64 {
		return v.I(`GeneratedRandomSeed`).(int64)
	}

	s.Let(`GeneratedRandomSeed`, func(v *testcase.V) interface{} {
		return time.Now().Unix()
	})

	manager := func(v *testcase.V) *rollouts.RolloutManager {
		return &rollouts.RolloutManager{
			Storage: GetStorage(v),
			RandSeedGenerator: func() int64 {
				return GetGeneratedRandomSeed(v)
			},
		}
	}

	s.Before(func(t *testing.T, v *testcase.V) {
		require.Nil(t, GetStorage(v).Truncate(rollouts.FeatureFlag{}))
		require.Nil(t, GetStorage(v).Truncate(rollouts.Pilot{}))
	})

	s.Describe(`SetPilotEnrollmentForFeature`, func(s *testcase.Spec) {

		GetNewEnrollment := func(v *testcase.V) bool {
			return v.I(`NewEnrollment`).(bool)
		}

		subject := func(v *testcase.V) error {
			return manager(v).SetPilotEnrollmentForFeature(
				GetFeatureFlagName(v),
				GetExternalPilotID(v),
				GetNewEnrollment(v),
			)
		}

		s.Let(`NewEnrollment`, func(v *testcase.V) interface{} {
			return rand.Intn(2) == 0
		})

		findFlag := func(v *testcase.V) *rollouts.FeatureFlag {
			iter := GetStorage(v).FindAll(&rollouts.FeatureFlag{})
			require.NotNil(v.T(), iter)
			require.True(v.T(), iter.Next())
			var ff rollouts.FeatureFlag
			require.Nil(v.T(), iter.Decode(&ff))
			require.False(v.T(), iter.Next())
			require.Nil(v.T(), iter.Err())
			return &ff
		}

		s.When(`no feature flag is seen ever before`, func(s *testcase.Spec) {
			s.Before(func(t *testing.T, v *testcase.V) {
				require.Nil(t, GetStorage(v).Truncate(rollouts.FeatureFlag{}))
			})

			s.Then(`feature flag created`, func(t *testing.T, v *testcase.V) {
				require.Nil(t, subject(v))

				flag := findFlag(v)
				require.Equal(t, GetFeatureFlagName(v), flag.Name)
				require.Equal(t, "", flag.Rollout.Strategy.URL)
				require.Equal(t, 0, flag.Rollout.Strategy.Percentage)
				require.Equal(t, GetGeneratedRandomSeed(v), flag.Rollout.RandSeedSalt)
			})

			s.Then(`pilot is enrollment is set for the feature is set`, func(t *testing.T, v *testcase.V) {
				require.Nil(t, subject(v))

				flag := findFlag(v)
				pilot, err := GetStorage(v).FindFlagPilotByExternalPilotID(flag.ID, GetExternalPilotID(v))
				require.Nil(t, err)
				require.NotNil(t, pilot)
				require.Equal(t, GetNewEnrollment(v), pilot.Enrolled)
				require.Equal(t, GetExternalPilotID(v), pilot.ExternalID)
			})
		})

		s.When(`feature flag already configured`, func(s *testcase.Spec) {
			s.Before(func(t *testing.T, v *testcase.V) {
				require.Nil(t, GetStorage(v).Save(GetFeatureFlag(v)))
			})

			s.Then(`flag is will not be recreated`, func(t *testing.T, v *testcase.V) {
				require.Nil(t, subject(v))

				count, err := iterators.Count(GetStorage(v).FindAll(rollouts.FeatureFlag{}))
				require.Nil(t, err)
				require.Equal(t, 1, count)

				flag := findFlag(v)
				require.Equal(t, GetFeatureFlag(v), flag)
			})

			s.And(`pilot already exists`, func(s *testcase.Spec) {
				s.Before(func(t *testing.T, v *testcase.V) {
					require.Nil(t, GetStorage(v).Save(GetPilot(v)))
				})

				s.And(`and pilot is has the opposite enrollment status`, func(s *testcase.Spec) {
					s.Let(`PilotEnrollment`, func(v *testcase.V) interface{} {
						return !GetNewEnrollment(v)
					})

					s.Then(`the original pilot is updated to the new enrollment status`, func(t *testing.T, v *testcase.V) {
						require.Nil(t, subject(v))
						flag := findFlag(v)

						pilot, err := GetStorage(v).FindFlagPilotByExternalPilotID(flag.ID, GetExternalPilotID(v))
						require.Nil(t, err)

						require.NotNil(t, pilot)
						require.Equal(t, GetNewEnrollment(v), pilot.Enrolled)
						require.Equal(t, GetExternalPilotID(v), pilot.ExternalID)
						require.Equal(t, GetPilot(v), pilot)

						count, err := iterators.Count(GetStorage(v).FindAll(rollouts.Pilot{}))
						require.Nil(t, err)
						require.Equal(t, 1, count)
					})
				})

				s.And(`pilot already has the same enrollment status`, func(s *testcase.Spec) {
					s.Let(`PilotEnrollment`, func(v *testcase.V) interface{} {
						return GetNewEnrollment(v)
					})

					s.Then(`pilot remain the same`, func(t *testing.T, v *testcase.V) {

						require.Nil(t, subject(v))
						ff := findFlag(v)

						pilot, err := GetStorage(v).FindFlagPilotByExternalPilotID(ff.ID, GetExternalPilotID(v))
						require.Nil(t, err)

						require.NotNil(t, pilot)
						require.Equal(t, GetNewEnrollment(v), pilot.Enrolled)
						require.Equal(t, GetExternalPilotID(v), pilot.ExternalID)

						count, err := iterators.Count(GetStorage(v).FindAll(rollouts.Pilot{}))
						require.Nil(t, err)
						require.Equal(t, 1, count)

					})
				})
			})

		})
	})

	s.Describe(`UpdateFeatureFlagRolloutPercentage`, func(s *testcase.Spec) {
		GetNewRolloutPercentage := func(v *testcase.V) int {
			return v.I(`NewRolloutPercentage`).(int)
		}

		subject := func(v *testcase.V) error {
			return manager(v).UpdateFeatureFlagRolloutPercentage(GetFeatureFlagName(v), GetNewRolloutPercentage(v))
		}

		s.When(`percentage less than 0`, func(s *testcase.Spec) {
			s.Let(`NewRolloutPercentage`, func(v *testcase.V) interface{} { return -1 + (rand.Intn(1024) * -1) })

			s.Then(`it will report error`, func(t *testing.T, v *testcase.V) {
				require.Error(t, subject(v))
			})
		})

		s.When(`percentage greater than 100`, func(s *testcase.Spec) {
			s.Let(`NewRolloutPercentage`, func(v *testcase.V) interface{} { return 101 + rand.Intn(1024) })

			s.Then(`it will report error`, func(t *testing.T, v *testcase.V) {
				require.Error(t, subject(v))
			})
		})

		s.When(`percentage is a number between 0 and 100`, func(s *testcase.Spec) {
			s.Let(`NewRolloutPercentage`, func(v *testcase.V) interface{} { return rand.Intn(101) })

			s.And(`feature flag was undefined until now`, func(s *testcase.Spec) {
				s.Before(func(t *testing.T, v *testcase.V) {
					require.Nil(t, GetStorage(v).Truncate(rollouts.FeatureFlag{}))
				})

				s.Then(`feature flag entry created with the percentage`, func(t *testing.T, v *testcase.V) {
					require.Nil(t, subject(v))
					flag, err := GetStorage(v).FindByFlagName(GetFeatureFlagName(v))
					require.Nil(t, err)
					require.NotNil(t, flag)

					require.Equal(t, GetFeatureFlagName(v), flag.Name)
					require.Equal(t, "", flag.Rollout.Strategy.URL)
					require.Equal(t, GetNewRolloutPercentage(v), flag.Rollout.Strategy.Percentage)
					require.Equal(t, GetGeneratedRandomSeed(v), flag.Rollout.RandSeedSalt)
				})
			})

			s.And(`feature flag is already exist with a different percentage`, func(s *testcase.Spec) {
				s.Let(`RolloutPercentage`, func(v *testcase.V) interface{} {
					for {
						roll := rand.Intn(101)
						if roll != GetNewRolloutPercentage(v) {
							return roll
						}
					}
				})

				s.Before(func(t *testing.T, v *testcase.V) {
					require.Nil(t, GetStorage(v).Save(GetFeatureFlag(v)))
				})

				s.Then(`the same feature flag kept but updated to the new percentage`, func(t *testing.T, v *testcase.V) {
					require.Nil(t, subject(v))
					flag, err := GetStorage(v).FindByFlagName(GetFeatureFlagName(v))
					require.Nil(t, err)
					require.NotNil(t, flag)

					require.Equal(t, GetFeatureFlag(v).ID, flag.ID)
					require.Equal(t, GetFeatureFlagName(v), flag.Name)
					require.Equal(t, "", flag.Rollout.Strategy.URL)
					require.Equal(t, GetNewRolloutPercentage(v), flag.Rollout.Strategy.Percentage)
					require.Equal(t, GetRolloutSeedSalt(v), flag.Rollout.RandSeedSalt)
				})
			})
		})
	})
}
