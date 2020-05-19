package release_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/domains/release"
	. "github.com/toggler-io/toggler/testing"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/stretchr/testify/require"
)

func TestRolloutManager(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()
	SetUp(s)

	s.Let(`manager`, func(t *testcase.T) interface{} {
		return &release.RolloutManager{Storage: ExampleStorage(t)}
	})

	s.Describe(`CreateFeatureFlag`, SpecRolloutManagerCreateFeatureFlag)
	s.Describe(`UpdateFeatureFlag`, SpecRolloutManagerUpdateFeatureFlag)
	s.Describe(`DeleteFeatureFlag`, SpecRolloutManagerDeleteFeatureFlag)
	s.Describe(`ListFeatureFlags`, SpecRolloutManagerListFeatureFlags)

	s.Describe(`SetPilotEnrollmentForFeature`, SpecSetPilotEnrollmentForFeature)
	s.Describe(`UnsetPilotEnrollmentForFeature`, SpecUnsetPilotEnrollmentForFeature)
	s.Describe(`GetAllReleaseFlagStatesOfThePilot`, SpecRolloutManagerGetAllReleaseFlagStatesOfThePilot)
}

func SpecRolloutManagerCreateFeatureFlag(s *testcase.Spec) {
	subject := func(t *testcase.T) error {
		return manager(t).CreateFeatureFlag(GetContext(t), GetReleaseFlag(t, `flag`))
	}

	s.Let(`flag`, func(t *testcase.T) interface{} {
		flag := Create(release.Flag{}).(*release.Flag)
		t.Defer(ExampleStorage(t).DeleteByID, GetContext(t), release.Flag{}, flag.ID)
		return flag
	})

	s.Then(`on valid input the values persisted`, func(t *testcase.T) {
		require.Nil(t, subject(t))
		require.Equal(t, GetReleaseFlag(t, `flag`),
			FindStoredReleaseFlagByName(t, GetReleaseFlag(t, `flag`).Name))
	})

	s.When(`name is empty`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { GetReleaseFlag(t, `flag`).Name = `` })

		s.Then(`it will fail with invalid feature name`, func(t *testcase.T) {
			require.Equal(t, release.ErrNameIsEmpty, subject(t))
		})
	})

	s.When(`feature flag`, func(s *testcase.Spec) {
		s.Context(`is nil`, func(s *testcase.Spec) {
			s.Let(`flag`, func(t *testcase.T) interface{} { return nil })

			s.Then(`it will return error about it`, func(t *testcase.T) {
				require.Error(t, subject(t))
			})
		})

		s.Context(`was not stored until now`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { require.Empty(t, GetReleaseFlag(t, `flag`).ID) })

			s.Then(`it will be persisted`, func(t *testcase.T) {
				require.Nil(t, subject(t))
				require.Equal(t, GetReleaseFlag(t, `flag`),
					FindStoredReleaseFlagByName(t, GetReleaseFlag(t, `flag`).Name))
			})
		})

		s.Context(`had been persisted previously`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { require.Nil(t, subject(t)) })

			s.When(`the id is not referring to the existing one`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) { GetReleaseFlag(t, `flag`).ID = `` })

				s.Then(`it will report feature flag already exists error`, func(t *testcase.T) {
					require.Equal(t, release.ErrFlagAlreadyExist, subject(t))
				})
			})

			s.When(`the ID is set and pointing to an existing flag`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					flag := GetReleaseFlag(t, `flag`)
					require.NotEmpty(t, flag.ID)

					var stored release.Flag
					found, err := ExampleStorage(t).FindByID(context.Background(), &stored, flag.ID)
					require.Nil(t, err)
					require.True(t, found)
				})

				s.Then(`it will report invalid action error`, func(t *testcase.T) {
					require.Equal(t, release.ErrInvalidAction, subject(t))
				})
			})
		})
	})
}

func SpecRolloutManagerUpdateFeatureFlag(s *testcase.Spec) {
	subject := func(t *testcase.T) error {
		return manager(t).UpdateFeatureFlag(GetContext(t), ExampleReleaseFlag(t))
	}

	s.When(`flag content is invalid`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { ExampleReleaseFlag(t).Name = `` })

		s.Then(`it will yield error`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})

	s.When(`flag content is valid`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			ExampleReleaseFlag(t).Name = RandomName()
		})

		s.Then(`it will update the name`, func(t *testcase.T) {
			require.Nil(t, subject(t))

			var f release.Flag
			found, err := ExampleStorage(t).FindByID(GetContext(t), &f, ExampleReleaseFlag(t).ID)
			require.Nil(t, err)
			require.True(t, found)
			require.Equal(t, ExampleReleaseFlag(t), &f)
		})
	})
}

func SpecRolloutManagerDeleteFeatureFlag(s *testcase.Spec) {
	subject := func(t *testcase.T) error {
		flagID := t.I(`flag ID`).(string)
		t.Log(`flagID:`, flagID)
		return manager(t).DeleteFeatureFlag(GetContext(t), flagID)
	}

	s.When(`feature flag id is empty`, func(s *testcase.Spec) {
		s.Let(`flag ID`, func(t *testcase.T) interface{} { return `` })

		s.Then(`it will return error about it`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})

	s.When(`feature flag id is unknown by the storage`, func(s *testcase.Spec) {
		s.Let(`flag ID`, func(t *testcase.T) interface{} {
			// real id
			id := ExampleReleaseFlag(t).ID
			// but no longer present in the storage
			require.Nil(t, ExampleStorage(t).DeleteByID(GetContext(t), release.Flag{}, id))
			return id
		})

		s.Then(`it will return error about it`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})

	s.When(`it had been persisted previously`, func(s *testcase.Spec) {
		s.Let(`flag ID`, func(t *testcase.T) interface{} {
			return ExampleReleaseFlag(t).ID
		})

		s.Then(`flag will be deleted`, func(t *testcase.T) {
			id := t.I(`flag ID`).(string)
			require.Nil(t, subject(t))

			var flag release.Flag
			found, err := manager(t).Storage.FindByID(GetContext(t), &flag, id)
			require.Nil(t, err)
			require.False(t, found)
		})

		s.And(`there are pilots manually set for the feature`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { ExampleReleaseManualPilotEnrollment(t) })

			var getPilotCount = func(t *testcase.T) int {
				count, err := iterators.Count(ExampleStorage(t).FindAll(GetContext(t), release.ManualPilot{}))
				require.Nil(t, err)
				return count
			}

			s.Then(`it will remove the pilots as well for the feature`, func(t *testcase.T) {
				require.Nil(t, subject(t))
				require.Equal(t, 0, getPilotCount(t))
			})

			s.And(`if other flags have pilots enrolled as well`, func(s *testcase.Spec) {
				GivenWeHaveReleaseFlag(s, `oth-flag`)

				s.Around(func(t *testcase.T) func() {
					env := ExampleDeploymentEnvironment(t)
					othFlag := GetReleaseFlag(t, `oth-flag`)
					require.Nil(t, manager(t).SetPilotEnrollmentForFeature(GetContext(t), othFlag.ID, env.ID, fixtures.Random.String(), true))
					require.Nil(t, manager(t).SetPilotEnrollmentForFeature(GetContext(t), othFlag.ID, env.ID, fixtures.Random.String(), false))
					return func() {
						require.Nil(t, ExampleStorage(t).DeleteAll(GetContext(t), release.ManualPilot{}))
					}
				})

				s.Then(`they will be unaffected by the subject flag removal`, func(t *testcase.T) {
					require.Nil(t, subject(t))
					require.Equal(t, 2, getPilotCount(t))
				})
			})
		})
	})
}

func SpecRolloutManagerListFeatureFlags(s *testcase.Spec) {
	subject := func(t *testcase.T) ([]release.Flag, error) {
		return manager(t).ListFeatureFlags(GetContext(t))
	}

	onSuccess := func(t *testcase.T) []release.Flag {
		ffs, err := subject(t)
		require.Nil(t, err)
		return ffs
	}

	s.When(`features are in the system`, func(s *testcase.Spec) {
		GivenWeHaveReleaseFlag(s, `flag-1`)
		GivenWeHaveReleaseFlag(s, `flag-2`)
		GivenWeHaveReleaseFlag(s, `flag-3`)

		s.Before(func(t *testcase.T) {
			var flags []release.Flag
			flags = append(flags, *GetReleaseFlag(t, `flag-1`))
			flags = append(flags, *GetReleaseFlag(t, `flag-2`))
			flags = append(flags, *GetReleaseFlag(t, `flag-3`))
			t.Let(`expected-flags`, flags)
		})

		s.Then(`feature flags are returned`, func(t *testcase.T) {
			require.ElementsMatch(t, t.I(`expected-flags`).([]release.Flag), onSuccess(t))
		})
	})

	s.When(`no feature present in the system`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { require.Nil(t, ExampleStorage(t).DeleteAll(GetContext(t), release.Flag{})) })

		s.Then(`empty list returned`, func(t *testcase.T) {
			require.Empty(t, onSuccess(t))
		})
	})
}

func SpecSetPilotEnrollmentForFeature(s *testcase.Spec) {
	getNewEnrollment := func(t *testcase.T) bool {
		return t.I(`new enrollment`).(bool)
	}

	subject := func(t *testcase.T) error {
		return manager(t).SetPilotEnrollmentForFeature(GetContext(t),
			ExampleReleaseFlag(t).ID,
			ExampleDeploymentEnvironment(t).ID,
			ExampleExternalPilotID(t),
			getNewEnrollment(t),
		)
	}

	s.Let(`new enrollment`, func(t *testcase.T) interface{} {
		return fixtures.Random.Bool()
	})

	s.When(`no feature flag is seen ever before`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, ExampleStorage(t).DeleteByID(GetContext(t), release.Flag{}, ExampleReleaseFlag(t).ID))
		})

		s.Then(`error returned`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})

	s.When(`pilot already exists`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { ExampleReleaseManualPilotEnrollment(t) })

		s.And(`and pilot is has the opposite enrollment status`, func(s *testcase.Spec) {
			s.LetValue(`new enrollment`, true)
			AndExamplePilotManualParticipatingIsSetTo(s, false)

			s.Then(`the original pilot is updated to the new enrollment status`, func(t *testcase.T) {
				require.Nil(t, subject(t))
				flag := ExampleReleaseFlag(t)
				env := ExampleDeploymentEnvironment(t)

				pilot, err := ExampleStorage(t).FindReleaseManualPilotByExternalID(
					GetContext(t), flag.ID, env.ID, ExampleExternalPilotID(t))

				require.Nil(t, err)

				require.NotNil(t, pilot)
				require.Equal(t, getNewEnrollment(t), pilot.IsParticipating)
				require.Equal(t, ExampleExternalPilotID(t), pilot.ExternalID)

				expectedPilot := release.ManualPilot{
					FlagID:                  ExampleReleaseFlag(t).ID,
					DeploymentEnvironmentID: ExampleDeploymentEnvironment(t).ID,
					ExternalID:              ExampleExternalPilotID(t),
					IsParticipating:         getNewEnrollment(t),
				}

				actualPilot := *pilot
				actualPilot.ID = ``

				require.Equal(t, expectedPilot, actualPilot)

				count, err := iterators.Count(ExampleStorage(t).FindAll(context.Background(), release.ManualPilot{}))
				require.Nil(t, err)
				require.Equal(t, 1, count)
			})
		})

		s.And(`pilot already has the same enrollment status`, func(s *testcase.Spec) {
			s.Let(`PilotEnrollment`, func(t *testcase.T) interface{} {
				return getNewEnrollment(t)
			})

			s.Then(`pilot remain the same`, func(t *testcase.T) {

				require.Nil(t, subject(t))
				ff := ExampleReleaseFlag(t)

				pilot, err := ExampleStorage(t).FindReleaseManualPilotByExternalID(context.Background(), ff.ID, ExampleDeploymentEnvironment(t).ID, ExampleExternalPilotID(t))
				require.Nil(t, err)

				require.NotNil(t, pilot)
				require.Equal(t, getNewEnrollment(t), pilot.IsParticipating)
				require.Equal(t, ExampleExternalPilotID(t), pilot.ExternalID)

				count, err := iterators.Count(ExampleStorage(t).FindAll(context.Background(), release.ManualPilot{}))
				require.Nil(t, err)
				require.Equal(t, 1, count)

			})
		})
	})
}

func SpecUnsetPilotEnrollmentForFeature(s *testcase.Spec) {
	subject := func(t *testcase.T) error {
		return manager(t).UnsetPilotEnrollmentForFeature(GetContext(t),
			ExampleReleaseFlag(t).ID,
			ExampleDeploymentEnvironment(t).ID,
			ExampleExternalPilotID(t),
		)
	}

	s.When(`flag is not persisted`, func(s *testcase.Spec) {
		s.Let(ExampleReleaseFlagLetVar, func(t *testcase.T) interface{} {
			return Create(release.Flag{})
		})

		s.Then(`error returned`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})

	s.When(`flag already persisted`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { ExampleReleaseFlag(t) }) // eager load

		s.And(`pilot doesn't exist for the flag`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, ExampleStorage(t).DeleteAll(GetContext(t), release.ManualPilot{}))
			})

			s.Then(`it will return without any error`, func(t *testcase.T) {
				require.Nil(t, subject(t))
			})
		})

		s.And(`pilot already exists`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { ExampleReleaseManualPilotEnrollment(t) }) // eager load

			s.Then(`pilot manual enrollment will be removed`, func(t *testcase.T) {
				require.Nil(t, subject(t))
				flag := ExampleReleaseFlag(t)

				pilot, err := ExampleStorage(t).FindReleaseManualPilotByExternalID(context.Background(), flag.ID, "", ExampleExternalPilotID(t))
				require.Nil(t, err)
				require.Nil(t, pilot)
			})
		})
	})
}

func SpecRolloutManagerGetAllReleaseFlagStatesOfThePilot(s *testcase.Spec) {
	subject := func(t *testcase.T) (bool, error) {
		states, err := manager(t).GetAllReleaseFlagStatesOfThePilot(GetContext(t),
			ExampleExternalPilotID(t),
			*ExampleDeploymentEnvironment(t),
			ExampleReleaseFlag(t).Name,
		)
		if err != nil {
			return false, err
		}
		return states[ExampleReleaseFlag(t).Name], nil
	}

	s.When(`feature was never enabled before`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			flag := ExampleReleaseFlag(t)
			require.Nil(t, ExampleStorage(t).DeleteByID(GetContext(t), *flag, flag.ID))
		})

		s.Then(`it will tell that feature flag is not enabled`, func(t *testcase.T) {
			ok, err := subject(t)
			require.Nil(t, err)
			require.False(t, ok)
		})
	})

	s.When(`flag and rollout plan is present`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			ExampleReleaseFlag(t)
			ExampleReleaseRollout(t)
		})

		s.And(`by rollout plan`, func(s *testcase.Spec) {
			s.Context(`it is not participating`, func(s *testcase.Spec) {
				AndReleaseFlagRolloutPercentageIs(s, ExampleReleaseRolloutLetVar, 0)

				s.Then(`pilot is not enrolled for the feature`, func(t *testcase.T) {
					ok, err := subject(t)
					require.Nil(t, err)
					require.False(t, ok)
				})

				s.Context(`but manual pilot config force enroll the given pilot`, func(s *testcase.Spec) {
					AndExamplePilotManualParticipatingIsSetTo(s, true)

					s.Then(`pilot is enrolled for the feature`, func(t *testcase.T) {
						ok, err := subject(t)
						require.Nil(t, err)
						require.True(t, ok)
					})
				})
			})

			s.Context(`it is participating`, func(s *testcase.Spec) {
				AndReleaseFlagRolloutPercentageIs(s, ExampleReleaseRolloutLetVar, 100)

				s.Then(`pilot is enrolled for the feature`, func(t *testcase.T) {
					ok, err := subject(t)
					require.Nil(t, err)
					require.True(t, ok)
				})

				s.Context(`but manual pilot config prevents participation for the given pilot`, func(s *testcase.Spec) {
					AndExamplePilotManualParticipatingIsSetTo(s, false)

					s.Then(`pilot is enrolled for the feature`, func(t *testcase.T) {
						ok, err := subject(t)
						require.Nil(t, err)
						require.False(t, ok)
					})
				})
			})
		})
	})

	s.Test(`E2E with percentage based rollout definition`, func(t *testcase.T) {
		var tolerationPercentage int
		if testing.Short() {
			tolerationPercentage = 5
		} else {
			tolerationPercentage = 3
		}
		var samplingCount int
		if testing.Short() {
			samplingCount = 1000
		} else {
			samplingCount = 10000
		}
		extIDS := make([]string, 0, samplingCount)
		for i := 0; i < samplingCount; i++ {
			extIDS = append(extIDS, RandomExternalPilotID())
		}
		expectedEnrollMaxPercentage := rand.Intn(51) + 50
		if 100 < expectedEnrollMaxPercentage+tolerationPercentage {
			tolerationPercentage = 100 - expectedEnrollMaxPercentage
		}

		flag := ExampleReleaseFlag(t)
		env := ExampleDeploymentEnvironment(t)

		byPercentage := release.NewRolloutDecisionByPercentage()
		byPercentage.Percentage = expectedEnrollMaxPercentage
		rollout := release.Rollout{
			FlagID:                  flag.ID,
			DeploymentEnvironmentID: env.ID,
			Plan:                    byPercentage,
		}
		require.Nil(t, ExampleStorage(t).Create(GetContext(t), &rollout))
		t.Defer(func() { require.Nil(t, ExampleStorage(t).DeleteByID(GetContext(t), rollout, rollout.ID)) })

		/* start E2E test */

		var enrolled, rejected int

		t.Log(`given we use the constructor`)
		rolloutManager := manager(t)

		for _, extID := range extIDS {
			releaseEnrollmentStates, err := rolloutManager.GetAllReleaseFlagStatesOfThePilot(GetContext(t), extID, *env, flag.Name)
			require.Nil(t, err)

			isIn, ok := releaseEnrollmentStates[flag.Name]
			require.True(t, ok, `release flag is not present in the enrollment states`)

			if isIn {
				enrolled++
			} else {
				rejected++
			}
		}

		require.True(t, enrolled > 0, `no one enrolled? fishy`)

		t.Logf(`a little toleration is still accepted, as long in generally it is within the range (+%d%%)`, tolerationPercentage)
		maximumAcceptedEnrollmentPercentage := expectedEnrollMaxPercentage + tolerationPercentage
		minimumAcceptedEnrollmentPercentage := expectedEnrollMaxPercentage - tolerationPercentage

		t.Logf(`so the total percentage in this test that fulfil the requirements is %d%%`, maximumAcceptedEnrollmentPercentage)
		testRunResultPercentage := int(float32(enrolled) / float32(enrolled+rejected) * 100)

		t.Logf(`and the actual percentage is %d%%`, testRunResultPercentage)
		require.True(t, testRunResultPercentage <= maximumAcceptedEnrollmentPercentage)
		require.True(t, minimumAcceptedEnrollmentPercentage <= testRunResultPercentage)

	})
}

func GetGeneratedRandomSeed(t *testcase.T) int64 {
	return t.I(`GeneratedRandomSeed`).(int64)
}

func manager(t *testcase.T) *release.RolloutManager {
	return t.I(`manager`).(*release.RolloutManager)
}
