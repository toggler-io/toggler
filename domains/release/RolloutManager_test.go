package release_test

import (
	"math/rand"
	"testing"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/domains/release"
	sh "github.com/toggler-io/toggler/spechelper"

	"github.com/stretchr/testify/require"
)

func TestRolloutManager(t *testing.T) {
	s := sh.NewSpec(t)
	s.Parallel()

	s.Let(`manager`, func(t *testcase.T) interface{} {
		return &release.RolloutManager{Storage: sh.StorageGet(t)}
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
	var subject = func(t *testcase.T) error {
		return manager(t).CreateFeatureFlag(sh.ContextGet(t), sh.GetReleaseFlag(t, `flag`))
	}

	s.Let(`flag`, func(t *testcase.T) interface{} {
		flag := sh.NewFixtureFactory(t).Create(release.Flag{}).(release.Flag)
		sh.StorageGet(t) // eager load storage to ensure storage teardown is not after the defer
		// pass by func used to ensure that flag.ID is populated by the time delete by id is called
		t.Cleanup(func() { _ = sh.StorageGet(t).ReleaseFlag(sh.ContextGet(t)).DeleteByID(sh.ContextGet(t), flag.ID) })
		return &flag
	})

	s.Then(`on valid input the values persisted`, func(t *testcase.T) {
		require.Nil(t, subject(t))
		require.Equal(t, sh.GetReleaseFlag(t, `flag`),
			sh.FindStoredReleaseFlagByName(t, sh.GetReleaseFlag(t, `flag`).Name))
	})

	s.When(`name is empty`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { sh.GetReleaseFlag(t, `flag`).Name = `` })

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
			s.Before(func(t *testcase.T) { require.Empty(t, sh.GetReleaseFlag(t, `flag`).ID) })

			s.Then(`it will be persisted`, func(t *testcase.T) {
				require.Nil(t, subject(t))
				require.Equal(t, sh.GetReleaseFlag(t, `flag`),
					sh.FindStoredReleaseFlagByName(t, sh.GetReleaseFlag(t, `flag`).Name))
			})
		})

		s.Context(`had been persisted previously`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { require.Nil(t, subject(t)) })

			s.When(`the id is not referring to the existing one`, func(s *testcase.Spec) {
				s.Around(func(t *testcase.T) func() {
					ogID := sh.GetReleaseFlag(t, `flag`).ID
					sh.GetReleaseFlag(t, `flag`).ID = ``
					return func() { sh.StorageGet(t).ReleaseFlag(sh.ContextGet(t)).DeleteByID(sh.ContextGet(t), ogID) }
				})

				s.Then(`it will report feature flag already exists error`, func(t *testcase.T) {
					require.Equal(t, release.ErrFlagAlreadyExist, subject(t))
				})
			})

			s.When(`the ID is set and pointing to an existing flag`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					flag := sh.GetReleaseFlag(t, `flag`)
					require.NotEmpty(t, flag.ID)

					var stored release.Flag
					found, err := sh.StorageGet(t).ReleaseFlag(sh.ContextGet(t)).FindByID(sh.ContextGet(t), &stored, flag.ID)
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
	var subject = func(t *testcase.T) error {
		return manager(t).UpdateFeatureFlag(sh.ContextGet(t), sh.ExampleReleaseFlag(t))
	}

	s.When(`flag content is invalid`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { sh.ExampleReleaseFlag(t).Name = `` })

		s.Then(`it will yield error`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})

	s.When(`flag content is valid`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			sh.ExampleReleaseFlag(t).Name = fixtures.Random.String()
		})

		s.Then(`it will update the name`, func(t *testcase.T) {
			require.Nil(t, subject(t))

			var f release.Flag
			found, err := sh.StorageGet(t).ReleaseFlag(sh.ContextGet(t)).FindByID(sh.ContextGet(t), &f, sh.ExampleReleaseFlag(t).ID)
			require.Nil(t, err)
			require.True(t, found)
			require.Equal(t, sh.ExampleReleaseFlag(t), &f)
		})
	})
}

func SpecRolloutManagerDeleteFeatureFlag(s *testcase.Spec) {
	var subject = func(t *testcase.T) error {
		flagID := t.I(`flag ID`).(string)
		t.Log(`flagID:`, flagID)
		return manager(t).DeleteFeatureFlag(sh.ContextGet(t), flagID)
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
			id := sh.ExampleReleaseFlag(t).ID
			// but no longer present in the storage
			require.Nil(t, sh.StorageGet(t).ReleaseFlag(sh.ContextGet(t)).DeleteByID(sh.ContextGet(t), id))
			return id
		})

		s.Then(`it will return error about it`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})

	s.When(`it had been persisted previously`, func(s *testcase.Spec) {
		s.Let(`flag ID`, func(t *testcase.T) interface{} {
			return sh.ExampleReleaseFlag(t).ID
		})

		s.Then(`flag will be deleted`, func(t *testcase.T) {
			id := t.I(`flag ID`).(string)
			require.Nil(t, subject(t))

			var flag release.Flag
			found, err := manager(t).Storage.ReleaseFlag(sh.ContextGet(t)).FindByID(sh.ContextGet(t), &flag, id)
			require.Nil(t, err)
			require.False(t, found)
		})

		s.And(`there are pilots manually set for the feature`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { sh.ExampleReleaseManualPilotEnrollment(t) })

			var getPilotCount = func(t *testcase.T) int {
				count, err := iterators.Count(sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).FindAll(sh.ContextGet(t)))
				require.Nil(t, err)
				return count
			}

			s.Then(`it will remove the pilots as well for the feature`, func(t *testcase.T) {
				require.Nil(t, subject(t))
				require.Equal(t, 0, getPilotCount(t))
			})

			s.And(`if other flags have pilots enrolled as well`, func(s *testcase.Spec) {
				sh.GivenWeHaveReleaseFlag(s, `oth-flag`)

				s.Around(func(t *testcase.T) func() {
					env := sh.ExampleDeploymentEnvironment(t)
					othFlag := sh.GetReleaseFlag(t, `oth-flag`)
					require.Nil(t, manager(t).SetPilotEnrollmentForFeature(sh.ContextGet(t), othFlag.ID, env.ID, fixtures.Random.String(), true))
					require.Nil(t, manager(t).SetPilotEnrollmentForFeature(sh.ContextGet(t), othFlag.ID, env.ID, fixtures.Random.String(), false))
					return func() {
						require.Nil(t, sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).DeleteAll(sh.ContextGet(t)))
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
	var subject = func(t *testcase.T) ([]release.Flag, error) {
		return manager(t).ListFeatureFlags(sh.ContextGet(t))
	}

	onSuccess := func(t *testcase.T) []release.Flag {
		ffs, err := subject(t)
		require.Nil(t, err)
		return ffs
	}

	s.When(`features are in the system`, func(s *testcase.Spec) {
		sh.GivenWeHaveReleaseFlag(s, `flag-1`)
		sh.GivenWeHaveReleaseFlag(s, `flag-2`)
		sh.GivenWeHaveReleaseFlag(s, `flag-3`)

		s.Before(func(t *testcase.T) {
			var flags []release.Flag
			flags = append(flags, *sh.GetReleaseFlag(t, `flag-1`))
			flags = append(flags, *sh.GetReleaseFlag(t, `flag-2`))
			flags = append(flags, *sh.GetReleaseFlag(t, `flag-3`))
			t.Set(`expected-flags`, flags)
		})

		s.Then(`feature flags are returned`, func(t *testcase.T) {
			require.ElementsMatch(t, t.I(`expected-flags`).([]release.Flag), onSuccess(t))
		})
	})

	s.When(`no feature present in the system`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { require.Nil(t, sh.StorageGet(t).ReleaseFlag(sh.ContextGet(t)).DeleteAll(sh.ContextGet(t))) })

		s.Then(`empty list returned`, func(t *testcase.T) {
			require.Empty(t, onSuccess(t))
		})
	})
}

func SpecSetPilotEnrollmentForFeature(s *testcase.Spec) {
	getNewEnrollment := func(t *testcase.T) bool {
		return t.I(`new enrollment`).(bool)
	}

	var subject = func(t *testcase.T) error {
		return manager(t).SetPilotEnrollmentForFeature(sh.ContextGet(t),
			sh.ExampleReleaseFlag(t).ID,
			sh.ExampleDeploymentEnvironment(t).ID,
			sh.ExampleExternalPilotID(t),
			getNewEnrollment(t),
		)
	}

	s.Let(`new enrollment`, func(t *testcase.T) interface{} {
		return fixtures.Random.Bool()
	})

	s.When(`no feature flag is seen ever before`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, sh.StorageGet(t).ReleaseFlag(sh.ContextGet(t)).DeleteByID(sh.ContextGet(t), sh.ExampleReleaseFlag(t).ID))
		})

		s.Then(`error returned`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})

	s.When(`pilot already exists`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { sh.ExampleReleaseManualPilotEnrollment(t) })

		s.And(`pilot is has the opposite enrollment status`, func(s *testcase.Spec) {
			s.LetValue(`new enrollment`, true)
			sh.AndExamplePilotManualParticipatingIsSetTo(s, false)

			s.Then(`the original pilot is updated to the new enrollment status`, func(t *testcase.T) {
				require.Nil(t, subject(t))
				flag := sh.ExampleReleaseFlag(t)
				env := sh.ExampleDeploymentEnvironment(t)

				pilot, err := sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).FindByFlagEnvPublicID(
					sh.ContextGet(t), flag.ID, env.ID, sh.ExampleExternalPilotID(t))

				require.Nil(t, err)

				require.NotNil(t, pilot)
				require.Equal(t, getNewEnrollment(t), pilot.IsParticipating)
				require.Equal(t, sh.ExampleExternalPilotID(t), pilot.PublicID)

				expectedPilot := release.Pilot{
					FlagID:          sh.ExampleReleaseFlag(t).ID,
					EnvironmentID:   sh.ExampleDeploymentEnvironment(t).ID,
					PublicID:        sh.ExampleExternalPilotID(t),
					IsParticipating: getNewEnrollment(t),
				}

				actualPilot := *pilot
				actualPilot.ID = ``

				require.Equal(t, expectedPilot, actualPilot)

				count, err := iterators.Count(sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).FindAll(sh.ContextGet(t)))
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
				ff := sh.ExampleReleaseFlag(t)

				pilot, err := sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).FindByFlagEnvPublicID(sh.ContextGet(t), ff.ID, sh.ExampleDeploymentEnvironment(t).ID, sh.ExampleExternalPilotID(t))
				require.Nil(t, err)

				require.NotNil(t, pilot)
				require.Equal(t, getNewEnrollment(t), pilot.IsParticipating)
				require.Equal(t, sh.ExampleExternalPilotID(t), pilot.PublicID)

				count, err := iterators.Count(sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).FindAll(sh.ContextGet(t)))
				require.Nil(t, err)
				require.Equal(t, 1, count)

			})
		})
	})
}

func SpecUnsetPilotEnrollmentForFeature(s *testcase.Spec) {
	var subject = func(t *testcase.T) error {
		return manager(t).UnsetPilotEnrollmentForFeature(sh.ContextGet(t),
			sh.ExampleReleaseFlag(t).ID,
			sh.ExampleDeploymentEnvironment(t).ID,
			sh.ExampleExternalPilotID(t),
		)
	}

	s.When(`flag is not persisted`, func(s *testcase.Spec) {
		s.Let(sh.LetVarExampleReleaseFlag, func(t *testcase.T) interface{} {
			rf := sh.NewFixtureFactory(t).Create(release.Flag{}).(release.Flag)
			return &rf
		})

		s.Then(`error returned`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})

	s.When(`flag already persisted`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { sh.ExampleReleaseFlag(t) }) // eager load

		s.And(`pilot doesn't exist for the flag`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).DeleteAll(sh.ContextGet(t)))
			})

			s.Then(`it will return without any error`, func(t *testcase.T) {
				require.Nil(t, subject(t))
			})
		})

		s.And(`pilot already exists`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { sh.ExampleReleaseManualPilotEnrollment(t) }) // eager load

			s.Then(`pilot manual enrollment will be removed`, func(t *testcase.T) {
				require.Nil(t, subject(t))
				pilot, err := sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).FindByFlagEnvPublicID(
					sh.ContextGet(t),
					sh.ExampleReleaseFlag(t).ID,
					sh.ExampleDeploymentEnvironment(t).ID,
					sh.ExampleExternalPilotID(t),
				)
				require.Nil(t, err)
				require.Nil(t, pilot)
			})
		})
	})
}

func SpecRolloutManagerGetAllReleaseFlagStatesOfThePilot(s *testcase.Spec) {
	var subject = func(t *testcase.T) (bool, error) {
		states, err := manager(t).GetAllReleaseFlagStatesOfThePilot(sh.ContextGet(t),
			sh.ExampleExternalPilotID(t),
			*sh.ExampleDeploymentEnvironment(t),
			sh.ExampleReleaseFlag(t).Name,
		)
		if err != nil {
			return false, err
		}
		return states[sh.ExampleReleaseFlag(t).Name], nil
	}

	s.When(`feature was never enabled before`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			flag := sh.ExampleReleaseFlag(t)
			require.Nil(t, sh.StorageGet(t).ReleaseFlag(sh.ContextGet(t)).DeleteByID(sh.ContextGet(t), flag.ID))
		})

		s.Then(`it will tell that feature flag is not enabled`, func(t *testcase.T) {
			ok, err := subject(t)
			require.Nil(t, err)
			require.False(t, ok)
		})
	})

	s.When(`flag and rollout plan is present`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			sh.ExampleReleaseFlag(t)
			sh.ExampleReleaseRollout(t)
		})

		s.And(`by rollout plan`, func(s *testcase.Spec) {
			s.Context(`it is not participating`, func(s *testcase.Spec) {
				sh.AndReleaseFlagRolloutPercentageIs(s, sh.LetVarExampleReleaseRollout, 0)

				s.Then(`pilot is not enrolled for the feature`, func(t *testcase.T) {
					ok, err := subject(t)
					require.Nil(t, err)
					require.False(t, ok)
				})

				s.Context(`but manual pilot config force enroll the given pilot`, func(s *testcase.Spec) {
					sh.AndExamplePilotManualParticipatingIsSetTo(s, true)

					s.Then(`pilot is enrolled for the feature`, func(t *testcase.T) {
						ok, err := subject(t)
						require.Nil(t, err)
						require.True(t, ok)
					})
				})
			})

			s.Context(`it is participating`, func(s *testcase.Spec) {
				sh.AndReleaseFlagRolloutPercentageIs(s, sh.LetVarExampleReleaseRollout, 100)

				s.Then(`pilot is enrolled for the feature`, func(t *testcase.T) {
					ok, err := subject(t)
					require.Nil(t, err)
					require.True(t, ok)
				})

				s.Context(`but manual pilot config prevents participation for the given pilot`, func(s *testcase.Spec) {
					sh.AndExamplePilotManualParticipatingIsSetTo(s, false)

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
			extIDS = append(extIDS, fixtures.Random.StringN(100))
		}
		expectedEnrollMaxPercentage := rand.Intn(51) + 50
		if 100 < expectedEnrollMaxPercentage+tolerationPercentage {
			tolerationPercentage = 100 - expectedEnrollMaxPercentage
		}

		flag := sh.ExampleReleaseFlag(t)
		env := sh.ExampleDeploymentEnvironment(t)

		byPercentage := release.NewRolloutDecisionByPercentage()
		byPercentage.Percentage = expectedEnrollMaxPercentage
		rollout := release.Rollout{
			FlagID:                  flag.ID,
			DeploymentEnvironmentID: env.ID,
			Plan:                    byPercentage,
		}
		require.Nil(t, sh.StorageGet(t).ReleaseRollout(sh.ContextGet(t)).Create(sh.ContextGet(t), &rollout))
		t.Defer(func() { require.Nil(t, sh.StorageGet(t).ReleaseRollout(sh.ContextGet(t)).DeleteByID(sh.ContextGet(t), rollout.ID)) })

		/* start E2E test */

		var enrolled, rejected int

		t.Log(`given we use the constructor`)
		rolloutManager := manager(t)

		for _, extID := range extIDS {
			releaseEnrollmentStates, err := rolloutManager.GetAllReleaseFlagStatesOfThePilot(sh.ContextGet(t), extID, *env, flag.Name)
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

func manager(t *testcase.T) *release.RolloutManager {
	return t.I(`manager`).(*release.RolloutManager)
}
