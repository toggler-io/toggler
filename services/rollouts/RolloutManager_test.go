package rollouts_test

import (
	"math/rand"
	"net/url"
	"testing"
	"time"

	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/testcase"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/stretchr/testify/require"
)

func TestRolloutManager(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()
	SetupSpecCommonVariables(s)

	s.Let(`GeneratedRandomSeed`, func(t *testcase.T) interface{} {
		return time.Now().Unix()
	})

	s.Let(`RolloutManager`, func(t *testcase.T) interface{} {
		return &rollouts.RolloutManager{
			Storage: GetStorage(t),

			RandSeedGenerator: func() int64 {
				return GetGeneratedRandomSeed(t)
			},
		}
	})

	s.Before(func(t *testcase.T) {
		require.Nil(t, GetStorage(t).Truncate(rollouts.FeatureFlag{}))
		require.Nil(t, GetStorage(t).Truncate(rollouts.Pilot{}))
	})

	SpecRolloutManagerCreateFeatureFlag(s)
	SpecRolloutManagerUpdateFeatureFlag(s)
	SpecRolloutManagerListFeatureFlags(s)
	SpecRolloutManagerSetPilotEnrollmentForFeature(s)
}

func SpecRolloutManagerCreateFeatureFlag(s *testcase.Spec) {
	s.Describe(`CreateFeatureFlag`, func(s *testcase.Spec) {
		subjectWithArgs := func(t *testcase.T, f *rollouts.FeatureFlag) error {
			return manager(t).CreateFeatureFlag(f)
		}

		subject := func(t *testcase.T) error {
			return subjectWithArgs(t, GetFeatureFlag(t))
		}

		s.Let(`FeatureFlagName`, func(t *testcase.T) interface{} { return ExampleFeatureName() })
		s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} { return nil })
		s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return rand.Intn(101) })
		s.Let(`RolloutSeedSalt`, func(t *testcase.T) interface{} { return int64(42) })
		s.Let(`FeatureFlag`, func(t *testcase.T) interface{} {
			ff := &rollouts.FeatureFlag{Name: t.I(`FeatureName`).(string)}
			ff.Rollout.RandSeedSalt = t.I(`RolloutSeedSalt`).(int64)
			ff.Rollout.Strategy.Percentage = t.I(`RolloutPercentage`).(int)
			ff.Rollout.Strategy.DecisionLogicAPI = GetRolloutApiURL(t)
			return ff
		})

		s.Then(`on valid input the values persisted`, func(t *testcase.T) {
			require.Nil(t, subject(t))
			require.NotNil(t, FindStoredFeatureFlagByName(t))
			require.Equal(t, GetFeatureFlag(t), FindStoredFeatureFlagByName(t))
		})

		s.When(`name is empty`, func(s *testcase.Spec) {
			s.Let(`FeatureName`, func(t *testcase.T) interface{} { return "" })

			s.Then(`it will fail with invalid feature name`, func(t *testcase.T) {
				require.Equal(t, rollouts.ErrNameIsEmpty, subject(t))
			})
		})

		s.When(`url`, func(s *testcase.Spec) {
			s.Context(`is not a valid request url`, func(s *testcase.Spec) {
				s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} { return `http//example.com` })

				s.Then(`it will fail with invalid url`, func(t *testcase.T) {
					require.Equal(t, rollouts.ErrInvalidRequestURL, subject(t))
				})
			})

			s.Context(`is not defined or nil`, func(s *testcase.Spec) {
				s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} { return nil })

				s.Then(`it will be saved and will represent that no custom domain decision url used`, func(t *testcase.T) {
					require.Nil(t, subject(t))

					require.Nil(t, FindStoredFeatureFlagByName(t).Rollout.Strategy.DecisionLogicAPI)
				})
			})

			s.Context(`is a valid request URL`, func(s *testcase.Spec) {
				s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} { return `https://example.com` })

				s.Then(`it will persist for future usage`, func(t *testcase.T) {
					require.Nil(t, subject(t))

					require.Equal(t, `https://example.com`, FindStoredFeatureFlagByName(t).Rollout.Strategy.DecisionLogicAPI.String())
				})
			})
		})

		s.When(`percentage`, func(s *testcase.Spec) {
			s.Context(`less than 0`, func(s *testcase.Spec) {
				s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return -1 + (rand.Intn(1024) * -1) })

				s.Then(`it will report error regarding the percentage`, func(t *testcase.T) {
					require.Equal(t, rollouts.ErrInvalidPercentage, subject(t))
				})
			})

			s.Context(`greater than 100`, func(s *testcase.Spec) {
				s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return 101 + rand.Intn(1024) })

				s.Then(`it will report error regarding the percentage`, func(t *testcase.T) {
					require.Equal(t, rollouts.ErrInvalidPercentage, subject(t))
				})
			})

			s.Context(`is a number between 0 and 100`, func(s *testcase.Spec) {
				s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return rand.Intn(101) })

				s.Then(`it will persist the received percentage`, func(t *testcase.T) {
					require.Nil(t, subject(t))

					require.Equal(t, t.I(`RolloutPercentage`).(int), FindStoredFeatureFlagByName(t).Rollout.Strategy.Percentage)
				})
			})
		})

		s.When(`pseudo random seed salt`, func(s *testcase.Spec) {
			s.Context(`is 0`, func(s *testcase.Spec) {
				s.Let(`RolloutSeedSalt`, func(t *testcase.T) interface{} { return int64(0) })

				s.Then(`random seed generator used for setting seed value`, func(t *testcase.T) {
					require.Nil(t, subject(t))

					require.Equal(t, GetGeneratedRandomSeed(t), FindStoredFeatureFlagByName(t).Rollout.RandSeedSalt)
				})
			})

			s.Context(`something else`, func(s *testcase.Spec) {
				s.Let(`RolloutSeedSalt`, func(t *testcase.T) interface{} { return int64(12) })

				s.Then(`it will persist the value`, func(t *testcase.T) {
					require.Nil(t, subject(t))

					require.Equal(t, int64(12), FindStoredFeatureFlagByName(t).Rollout.RandSeedSalt)
				})
			})
		})

		s.When(`feature flag`, func(s *testcase.Spec) {
			s.Context(`is nil`, func(s *testcase.Spec) {
				s.Let(`FeatureFlag`, func(t *testcase.T) interface{} { return nil })

				s.Then(`it will return error about it`, func(t *testcase.T) {
					require.Error(t, subject(t))
				})
			})

			s.Context(`was not stored until now`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, GetStorage(t).Truncate(rollouts.FeatureFlag{}))
				})

				s.Then(`it will be persisted`, func(t *testcase.T) {
					require.Nil(t, subject(t))
					require.NotNil(t, FindStoredFeatureFlagByName(t))
					require.Equal(t, GetFeatureFlag(t), FindStoredFeatureFlagByName(t))
				})
			})

			s.Context(`had been persisted previously`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, GetStorage(t).Save(GetFeatureFlag(t)))
					require.NotEmpty(t, GetFeatureFlag(t).ID)
				})

				s.When(`the id is not referring to the existing one`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						GetFeatureFlag(t).ID = ``
					})

					s.Then(`it will report feature flag already exists error`, func(t *testcase.T) {
						require.Equal(t, rollouts.ErrFlagAlreadyExist, subject(t))
					})
				})

				s.When(`the ID is set and pointing to an existing flag`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						require.NotEmpty(t, GetFeatureFlag(t).ID)
						var ff rollouts.FeatureFlag
						found, err := GetStorage(t).FindByID(GetFeatureFlag(t).ID, &ff)
						require.Nil(t, err)
						require.True(t, found)
						require.Equal(t, GetFeatureFlag(t), &ff)
					})

					s.Then(`it will report invalid action error`, func(t *testcase.T) {
						require.Equal(t, rollouts.ErrInvalidAction, subject(t))
					})
				})
			})
		})
	})
}

func SpecRolloutManagerUpdateFeatureFlag(s *testcase.Spec) {
	s.Describe(`UpdateFeatureFlag`, func(s *testcase.Spec) {
		subjectWithArgs := func(t *testcase.T, f *rollouts.FeatureFlag) error {
			return manager(t).UpdateFeatureFlag(f)
		}

		subject := func(t *testcase.T) error {
			return subjectWithArgs(t, GetFeatureFlag(t))
		}

		s.Let(`FeatureFlagName`, func(t *testcase.T) interface{} { return ExampleFeatureName() })
		s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} { return nil })
		s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return rand.Intn(101) })
		s.Let(`RolloutSeedSalt`, func(t *testcase.T) interface{} { return int64(42) })
		s.Let(`FeatureFlag`, func(t *testcase.T) interface{} {
			ff := &rollouts.FeatureFlag{Name: t.I(`FeatureName`).(string)}
			ff.Rollout.RandSeedSalt = t.I(`RolloutSeedSalt`).(int64)
			ff.Rollout.Strategy.Percentage = t.I(`RolloutPercentage`).(int)
			ff.Rollout.Strategy.DecisionLogicAPI = GetRolloutApiURL(t)
			return ff
		})

		s.When(`input is invalid for the feature flag Verify low level domain requirement`, func(s *testcase.Spec) {
			s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return 128 })

			s.Then(`it will report error`, func(t *testcase.T) {
				require.Error(t, subject(t))
			})
		})

		s.When(`feature flag`, func(s *testcase.Spec) {
			s.Context(`is nil`, func(s *testcase.Spec) {
				s.Let(`FeatureFlag`, func(t *testcase.T) interface{} { return nil })

				s.Then(`it will return error about it`, func(t *testcase.T) {
					require.Error(t, subject(t))
				})
			})

			s.Context(`was not stored until now`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, GetStorage(t).Truncate(rollouts.FeatureFlag{}))
				})

				s.Then(`it will report error about the missing ID`, func(t *testcase.T) {
					require.Error(t, subject(t))
				})
			})

			s.Context(`had been persisted previously`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, GetStorage(t).Save(GetFeatureFlag(t)))
					require.NotEmpty(t, GetFeatureFlag(t).ID)
				})

				s.Then(`latest values are persisted in the storage`, func(t *testcase.T) {
					flag := *GetFeatureFlag(t) // pass by value copy
					newName := flag.Name + ` v2`
					flag.Name = newName
					flag.Rollout.Strategy.Percentage = 42
					u, err := url.Parse(`https://example.com`)
					require.Nil(t, err)
					flag.Rollout.Strategy.DecisionLogicAPI = u
					require.Nil(t, subjectWithArgs(t, &flag))

					var storedFlag rollouts.FeatureFlag
					found, err := GetStorage(t).FindByID(GetFeatureFlag(t).ID, &storedFlag)
					require.Nil(t, err)
					require.True(t, found)
					require.Equal(t, u, storedFlag.Rollout.Strategy.DecisionLogicAPI)
					require.Equal(t, 42, storedFlag.Rollout.Strategy.Percentage)
					require.Equal(t, newName, storedFlag.Name)
				})
			})
		})

	})
}

func SpecRolloutManagerListFeatureFlags(s *testcase.Spec) {
	s.Describe(`ListFeatureFlags`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) ([]*rollouts.FeatureFlag, error) {
			return manager(t).ListFeatureFlags()
		}

		onSuccess := func(t *testcase.T) []*rollouts.FeatureFlag {
			ffs, err := subject(t)
			require.Nil(t, err)
			return ffs
		}

		s.When(`features are in the system`, func(s *testcase.Spec) {

			s.Before(func(t *testcase.T) {
				EnsureFlag(t, `a`, 0)
				EnsureFlag(t, `b`, 0)
				EnsureFlag(t, `c`, 0)
			})

			s.Then(`feature flags are returned`, func(t *testcase.T) {
				flags := onSuccess(t)

				expectedFlagNames := []string{`a`, `b`, `c`}

				for _, ff := range flags {
					require.Contains(t, expectedFlagNames, ff.Name)
				}
			})

		})

		s.When(`no feature present in the system`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, GetStorage(t).Truncate(rollouts.FeatureFlag{}))
			})

			s.Then(`feature flags are returned`, func(t *testcase.T) {
				flags := onSuccess(t)

				require.Equal(t, []*rollouts.FeatureFlag{}, flags)
			})
		})

	})
}

func SpecRolloutManagerSetPilotEnrollmentForFeature(s *testcase.Spec) {
	s.Describe(`SetPilotEnrollmentForFeature`, func(s *testcase.Spec) {

		GetNewEnrollment := func(t *testcase.T) bool {
			return t.I(`NewEnrollment`).(bool)
		}

		subject := func(t *testcase.T) error {
			return manager(t).SetPilotEnrollmentForFeature(
				t.I(`FeatureFlagID`).(string),
				GetExternalPilotID(t),
				GetNewEnrollment(t),
			)
		}

		s.Let(`FeatureFlagID`, func(t *testcase.T) interface{} {
			return GetFeatureFlag(t).ID
		})

		s.Let(`NewEnrollment`, func(t *testcase.T) interface{} {
			return rand.Intn(2) == 0
		})

		findFlag := func(t *testcase.T) *rollouts.FeatureFlag {
			iter := GetStorage(t).FindAll(&rollouts.FeatureFlag{})
			require.NotNil(t, iter)
			require.True(t, iter.Next())
			var ff rollouts.FeatureFlag
			require.Nil(t, iter.Decode(&ff))
			require.False(t, iter.Next())
			require.Nil(t, iter.Err())
			return &ff
		}

		s.When(`no feature flag is seen ever before`, func(s *testcase.Spec) {
			s.Let(`FeatureFlagID`, func(t *testcase.T) interface{} { return `` })
			s.Before(func(t *testcase.T) {
				require.Nil(t, GetStorage(t).Truncate(rollouts.FeatureFlag{}))
			})

			s.Then(`error returned`, func(t *testcase.T) {
				require.Error(t, subject(t))
			})
		})

		s.When(`feature flag already configured`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, GetStorage(t).Save(GetFeatureFlag(t)))
			})

			s.Then(`flag is will not be recreated`, func(t *testcase.T) {
				require.Nil(t, subject(t))

				count, err := iterators.Count(GetStorage(t).FindAll(rollouts.FeatureFlag{}))
				require.Nil(t, err)
				require.Equal(t, 1, count)

				flag := findFlag(t)
				require.Equal(t, GetFeatureFlag(t), flag)
			})

			s.And(`pilot already exists`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, GetStorage(t).Save(GetPilot(t)))
				})

				s.And(`and pilot is has the opposite enrollment status`, func(s *testcase.Spec) {
					s.Let(`PilotEnrollment`, func(t *testcase.T) interface{} {
						return !GetNewEnrollment(t)
					})

					s.Then(`the original pilot is updated to the new enrollment status`, func(t *testcase.T) {
						require.Nil(t, subject(t))
						flag := findFlag(t)

						pilot, err := GetStorage(t).FindFlagPilotByExternalPilotID(flag.ID, GetExternalPilotID(t))
						require.Nil(t, err)

						require.NotNil(t, pilot)
						require.Equal(t, GetNewEnrollment(t), pilot.Enrolled)
						require.Equal(t, GetExternalPilotID(t), pilot.ExternalID)
						require.Equal(t, GetPilot(t), pilot)

						count, err := iterators.Count(GetStorage(t).FindAll(rollouts.Pilot{}))
						require.Nil(t, err)
						require.Equal(t, 1, count)
					})
				})

				s.And(`pilot already has the same enrollment status`, func(s *testcase.Spec) {
					s.Let(`PilotEnrollment`, func(t *testcase.T) interface{} {
						return GetNewEnrollment(t)
					})

					s.Then(`pilot remain the same`, func(t *testcase.T) {

						require.Nil(t, subject(t))
						ff := findFlag(t)

						pilot, err := GetStorage(t).FindFlagPilotByExternalPilotID(ff.ID, GetExternalPilotID(t))
						require.Nil(t, err)

						require.NotNil(t, pilot)
						require.Equal(t, GetNewEnrollment(t), pilot.Enrolled)
						require.Equal(t, GetExternalPilotID(t), pilot.ExternalID)

						count, err := iterators.Count(GetStorage(t).FindAll(rollouts.Pilot{}))
						require.Nil(t, err)
						require.Equal(t, 1, count)

					})
				})
			})
		})
	})
}

func GetGeneratedRandomSeed(t *testcase.T) int64 {
	return t.I(`GeneratedRandomSeed`).(int64)
}

func manager(t *testcase.T) *rollouts.RolloutManager {
	return t.I(`RolloutManager`).(*rollouts.RolloutManager)
}
