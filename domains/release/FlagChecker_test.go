package release_test

import (
	"context"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	. "github.com/toggler-io/toggler/testing"
)

func TestFeatureFlagChecker(t *testing.T) {
	t.Parallel()

	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	s.Parallel()

	s.Let(`PseudoRandPercentage`, func(t *testcase.T) interface{} { return int(0) })

	s.Before(func(t *testcase.T) {
		require.Nil(t, GetStorage(t).Truncate(context.Background(), release.Flag{}))
		require.Nil(t, GetStorage(t).Truncate(context.Background(), release.Pilot{}))
	})

	s.Describe(`IsFeatureGloballyEnabled`, func(s *testcase.Spec) {
		SpecFeatureFlagChecker_IsFeatureGloballyEnabledFor(s)
	})

	s.Describe(`GetReleaseFlagPilotEnrollmentStates`, func(s *testcase.Spec) {
		SpecFeatureFlagChecker_GetReleaseFlagPilotEnrollmentStates(s)
		SpecFeatureFlagChecker_GetReleaseFlagPilotEnrollmentStates_BackwardCompatiblityWithIsFeatureEnabledFunctionalities(s)
	})
}

func SpecFeatureFlagChecker_GetReleaseFlagPilotEnrollmentStates(s *testcase.Spec) {
	subject := func(t *testcase.T) (map[string]bool, error) {
		return featureFlagChecker(t).GetReleaseFlagPilotEnrollmentStates(
			t.I(`ctx`).(context.Context),
			GetExternalPilotID(t),
			t.I(`flag names`).([]string)...)
	}

	s.Let(`ctx`, func(t *testcase.T) interface{} {
		return context.Background()
	})

	stateIs := func(t testing.TB, states map[string]bool, key string, expectedValue bool) {
		actual, ok := states[key]
		require.True(t, ok)
		require.Equal(t, expectedValue, actual)
	}

	cleanup := func(t *testcase.T) {
		ctx := context.Background()
		require.Nil(t, GetStorage(t).Truncate(ctx, release.Pilot{}))
		require.Nil(t, GetStorage(t).Truncate(ctx, release.Flag{}))
	}

	s.Before(cleanup)
	s.After(cleanup)

	s.When(`no flag is set`, func(s *testcase.Spec) {
		s.Let(`flag names`, func(t *testcase.T) interface{} {
			return []string{GetReleaseFlagName(t)}
		})
		s.Before(func(t *testcase.T) {
			f, err := GetStorage(t).FindReleaseFlagByName(context.Background(), GetReleaseFlagName(t))
			require.Nil(t, err)
			require.Nil(t, f, `there should be no release flag in the storage`)
		})

		s.Then(`it will return with flipped off states`, func(t *testcase.T) {
			states, err := subject(t)
			require.Nil(t, err)
			stateIs(t, states, GetReleaseFlagName(t), false)
		})
	})

	s.When(`flag exists`, func(s *testcase.Spec) {
		s.Let(`flag names`, func(t *testcase.T) interface{} {
			return []string{GetReleaseFlagName(t)}
		})

		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Create(context.Background(), GetReleaseFlag(t)))
		})

		// fix the value, so the dependency sub context can work with an assumption
		s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return int(42) })

		s.And(`there a non existing flag is also requested`, func(s *testcase.Spec) {
			s.Let(`non-existent-flag-name`, func(t *testcase.T) interface{} {
				return `non-existent`
			})
			s.Let(`flag names`, func(t *testcase.T) interface{} {
				return []string{GetReleaseFlagName(t), t.I(`non-existent-flag-name`).(string)}
			})

			s.Then(`flag state returned even for the existing one`, func(t *testcase.T) {
				states, err := subject(t)
				require.Nil(t, err)
				require.Equal(t, 2, len(states))
				_, ok := states[GetReleaseFlagName(t)]
				require.True(t, ok, `flag state should be returned`)
			})

			s.Then(`flag state returned even for the non-existing flag`, func(t *testcase.T) {
				states, err := subject(t)
				require.Nil(t, err)
				require.Equal(t, 2, len(states))
				enrolled, ok := states[t.I(`non-existent-flag-name`).(string)]
				require.True(t, ok)
				require.False(t, enrolled, `pilot should not be able to participant in a non existing release`)
			})
		})

		s.And(`the pilot would win the enrollment dice roll`, func(s *testcase.Spec) {
			s.Let(`PseudoRandPercentage`, func(t *testcase.T) interface{} { return int(0) })

			s.Then(`it is expected to receive switched ON state for the flag`, func(t *testcase.T) {
				states, err := subject(t)
				require.Nil(t, err)
				stateIs(t, states, GetReleaseFlagName(t), true)
			})

			s.And(`manually removed from the flag`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, GetStorage(t).Create(context.Background(), &release.Pilot{
						ExternalID: GetExternalPilotID(t),
						FlagID:     GetReleaseFlag(t).ID,
						Enrolled:   false,
					}))
				})

				s.Then(`the pilot flag state will be set to OFF`, func(t *testcase.T) {
					states, err := subject(t)
					require.Nil(t, err)
					stateIs(t, states, GetReleaseFlagName(t), false)
				})
			})
		})

		s.And(`the pilot would lose the enrollment dice roll`, func(s *testcase.Spec) {
			s.Let(`PseudoRandPercentage`, func(t *testcase.T) interface{} { return int(100) })

			s.Then(`it is expected to receive switched OFF state for the flag`, func(t *testcase.T) {
				states, err := subject(t)
				require.Nil(t, err)
				stateIs(t, states, GetReleaseFlagName(t), false)
			})

			s.And(`the pilot is white listed`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, GetStorage(t).Create(context.Background(), &release.Pilot{
						ExternalID: GetExternalPilotID(t),
						FlagID:     GetReleaseFlag(t).ID,
						Enrolled:   true,
					}))
				})

				s.Then(`the pilot flag state will be set to ON`, func(t *testcase.T) {
					states, err := subject(t)
					require.Nil(t, err)
					stateIs(t, states, GetReleaseFlagName(t), true)
				})
			})

			s.And(`the context has value about the pilot ip address`, func(s *testcase.Spec) {
				s.Let(`ip-addr`, func(t *testcase.T) interface{} {
					return `192.168.1.42`
				})
				s.Let(`ctx`, func(t *testcase.T) interface{} {
					return context.WithValue(context.Background(), `pilot-ip-addr`, t.I(`ip-addr`).(string))
				})

				s.And(`the ip address is configured to be allowed for the release`, func(s *testcase.Spec) {
					s.Around(func(t *testcase.T) func() {
						// TODO move it into the rollout manager logic
						a := &release.IPAllow{FlagID: GetReleaseFlag(t).ID, InternetProtocolAddress: t.I(`ip-addr`).(string)}
						require.Nil(t, GetStorage(t).Create(context.Background(), a))
						return func() {
							require.Nil(t, GetStorage(t).DeleteByID(context.Background(), release.IPAllow{}, a.ID))
						}
					})

					s.Then(`the pilot is enrolled`, func(t *testcase.T) {
						states, err := subject(t)
						require.Nil(t, err)
						require.True(t, states[GetReleaseFlagName(t)])
					})
				})
			})
		})

		s.Then(`it will always return state for the requested flag`, func(t *testcase.T) {
			states, err := subject(t)
			require.Nil(t, err)
			_, ok := states[GetReleaseFlagName(t)]
			require.True(t, ok, `flag state should be returned`)
		})
	})

	s.Test(`E2E`, func(t *testcase.T) {
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
			extIDS = append(extIDS, ExampleExternalPilotID())
		}
		expectedEnrollMaxPercentage := rand.Intn(51) + 50
		if 100 < expectedEnrollMaxPercentage+tolerationPercentage {
			tolerationPercentage = 100 - expectedEnrollMaxPercentage
		}
		releaseFlagName := GetReleaseFlagName(t)
		flag := &release.Flag{
			Name: releaseFlagName,
			Rollout: release.FlagRollout{
				Strategy: release.FlagRolloutStrategy{
					Percentage: expectedEnrollMaxPercentage,
				},
			},
		}
		require.Nil(t, release.NewRolloutManager(GetStorage(t)).CreateFeatureFlag(context.Background(), flag))
		defer GetStorage(t).DeleteByID(context.Background(), release.Flag{}, flag.ID)

		/* start E2E test */

		var enrolled, rejected int

		t.Log(`given we use the constructor`)
		ffc := release.NewFlagChecker(GetStorage(t))

		for _, extID := range extIDS {
			releaseEnrollmentStates, err := ffc.GetReleaseFlagPilotEnrollmentStates(context.Background(), extID, releaseFlagName)
			require.Nil(t, err)

			isIn, ok := releaseEnrollmentStates[releaseFlagName]
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

// TODO: merge the business specification from this into the main spec suite of the GetReleaseFlagPilotEnrollmentStates
func SpecFeatureFlagChecker_GetReleaseFlagPilotEnrollmentStates_BackwardCompatiblityWithIsFeatureEnabledFunctionalities(s *testcase.Spec) {
	subject := func(t *testcase.T) (bool, error) {
		releaseFlagName := t.I(`ReleaseFlagName`).(string)
		states, err := featureFlagChecker(t).GetReleaseFlagPilotEnrollmentStates(context.Background(), t.I(`PilotExternalID`).(string), releaseFlagName)
		if err != nil {
			return false, err
		}
		return states[releaseFlagName], nil
	}

	s.When(`feature was never enabled before`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Truncate(context.Background(), release.Flag{}))
		})

		s.Then(`it will tell that feature flag is not enabled`, func(t *testcase.T) {
			ok, err := subject(t)
			require.Nil(t, err)
			require.False(t, ok)
		})
	})

	s.When(`feature is configured with rollout strategy`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Create(context.TODO(), GetReleaseFlag(t)))
		})

		s.And(`by rollout percentage`, func(s *testcase.Spec) {
			s.And(`it is 0`, func(s *testcase.Spec) {
				s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return int(0) })

				s.And(`pseudo rand percentage for the id`, func(s *testcase.Spec) {
					s.And(`it is 0 as well`, func(s *testcase.Spec) {
						s.Let(`PseudoRandPercentage`, func(t *testcase.T) interface{} { return int(0) })

						s.Then(`pilot is not enrolled for the feature`, func(t *testcase.T) {
							ok, err := subject(t)
							require.Nil(t, err)
							require.False(t, ok)
						})
					})

					s.And(`it is greater than 0`, func(s *testcase.Spec) {
						s.Let(`PseudoRandPercentage`, func(t *testcase.T) interface{} { return int(42) })

						s.Then(`pilot is not enrolled for the feature`, func(t *testcase.T) {
							ok, err := subject(t)
							require.Nil(t, err)
							require.False(t, ok)
						})
					})
				})
			})

			s.And(`it is greater than 0`, func(s *testcase.Spec) {
				s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return rand.Intn(99) + 1 })

				s.And(`the next pseudo rand int is less or eq with the rollout percentage`, func(s *testcase.Spec) {
					s.Let(`PseudoRandPercentage`, func(t *testcase.T) interface{} {
						ffPerc := GetReleaseFlag(t).Rollout.Strategy.Percentage
						return ffPerc - rand.Intn(ffPerc)
					})

					s.Then(`it will enroll the pilot for the feature`, func(t *testcase.T) {
						ok, err := subject(t)
						require.Nil(t, err)
						require.True(t, ok)
					})

					s.And(`the pilot is blacklisted manually from the feature`, func(s *testcase.Spec) {
						s.Before(func(t *testcase.T) {

							pilot := &release.Pilot{
								FlagID:     GetReleaseFlag(t).ID,
								ExternalID: t.I(`PilotExternalID`).(string),
								Enrolled:   false,
							}

							require.Nil(t, GetStorage(t).Create(context.TODO(), pilot))

						})

						s.Then(`the pilot will be not enrolled for the feature flag`, func(t *testcase.T) {
							ok, err := subject(t)
							require.Nil(t, err)
							require.False(t, ok)
						})
					})

				})

				s.And(`the next pseudo rand int is is greater than the rollout percentage`, func(s *testcase.Spec) {

					s.Let(`PseudoRandPercentage`, func(t *testcase.T) interface{} {
						return GetReleaseFlag(t).Rollout.Strategy.Percentage + 1
					})

					s.Then(`pilot is not enrolled for the feature`, func(t *testcase.T) {
						ok, err := subject(t)
						require.Nil(t, err)
						require.False(t, ok)
					})

					s.And(`if pilot manually enabled for the feature`, func(s *testcase.Spec) {
						s.Let(`PilotEnrollment`, func(t *testcase.T) interface{} { return true })

						s.Before(func(t *testcase.T) {
							require.Nil(t, GetStorage(t).Create(context.TODO(), GetPilot(t)))
						})

						s.Then(`the pilot is enrolled for the feature`, func(t *testcase.T) {
							ok, err := subject(t)
							require.Nil(t, err)
							require.True(t, ok)
						})
					})
				})
			})

			s.And(`it is 100% or in other words it is set to be globally enabled`, func(s *testcase.Spec) {
				s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return int(100) })

				s.And(`basically regardless the pseudo random percentage`, func(s *testcase.Spec) {
					s.Let(`PseudoRandPercentage`, func(t *testcase.T) interface{} { return rand.Intn(101) })

					s.Then(`it will always be enrolled`, func(t *testcase.T) {
						ok, err := subject(t)
						require.Nil(t, err)
						require.True(t, ok)
					})
				})
			})

		})

		s.And(`by custom decision logic is defined with DecisionLogicAPI endpoint`, func(s *testcase.Spec) {
			s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} {
				return t.I(`httptest.NewServer`).(*httptest.Server).URL
			})

			handler := func(t *testcase.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, t.I(`ReleaseFlagName`).(string), r.URL.Query().Get(`feature`))
					require.Equal(t, t.I(`PilotExternalID`).(string), r.URL.Query().Get(`id`))
					w.WriteHeader(t.I(`replyCode`).(int))
				})
			}

			s.Let(`httptest.NewServer`, func(t *testcase.T) interface{} {
				return httptest.NewServer(handler(t))
			})

			s.After(func(t *testcase.T) {
				t.I(`httptest.NewServer`).(*httptest.Server).Close()
			})

			s.And(`the remote reject the pilot enrollment`, func(s *testcase.Spec) {
				s.Let(`replyCode`, func(t *testcase.T) interface{} { return rand.Intn(100) + 400 })

				s.Then(`the pilot is not enrolled for the feature`, func(t *testcase.T) {
					enabled, err := subject(t)
					require.Nil(t, err)
					require.False(t, enabled)
				})
			})

			s.And(`the the remote accept the pilot enrollment`, func(s *testcase.Spec) {
				s.Let(`replyCode`, func(t *testcase.T) interface{} { return rand.Intn(100) + 200 })

				s.Then(`the pilot is not enrolled for the feature`, func(t *testcase.T) {
					enabled, err := subject(t)
					require.Nil(t, err)
					require.True(t, enabled)
				})
			})
		})
	})
}

func SpecFeatureFlagChecker_IsFeatureGloballyEnabledFor(s *testcase.Spec) {
	subject := func(t *testcase.T) (bool, error) {
		return featureFlagChecker(t).IsFeatureGloballyEnabled(t.I(`ReleaseFlagName`).(string))
	}

	thenItWillReportThatFeatureNotGlobballyEnabled := func(s *testcase.Spec) {
		s.Then(`it will report that feature is not enabled globally`, func(t *testcase.T) {
			enabled, err := subject(t)
			require.Nil(t, err)
			require.False(t, enabled)
		})
	}

	thenItWillReportThatFeatureIsGloballyRolledOut := func(s *testcase.Spec) {
		s.Then(`it will report that feature is globally rolled out`, func(t *testcase.T) {
			enabled, err := subject(t)
			require.Nil(t, err)
			require.True(t, enabled)
		})
	}

	s.When(`feature flag is not seen before`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Truncate(context.Background(), release.Flag{}))
		})

		thenItWillReportThatFeatureNotGlobballyEnabled(s)
	})

	s.When(`feature flag is given`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Create(context.TODO(), GetReleaseFlag(t)))
		})

		s.And(`it is not yet rolled out globally`, func(s *testcase.Spec) {
			s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return 99 })

			thenItWillReportThatFeatureNotGlobballyEnabled(s)
		})

		s.And(`it is rolled out globally`, func(s *testcase.Spec) {
			s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return 100 })

			thenItWillReportThatFeatureIsGloballyRolledOut(s)
		})
	})
}

func featureFlagChecker(t *testcase.T) *release.FlagChecker {
	return &release.FlagChecker{
		Storage: t.I(`TestStorage`).(*TestStorage),
		IDPercentageCalculator: func(id string, seedSalt int64) (int, error) {
			require.Equal(t, t.I(`PilotExternalID`).(string), id)
			require.Equal(t, t.I(`RolloutSeedSalt`).(int64), seedSalt)
			return t.I(`PseudoRandPercentage`).(int), nil
		},
	}
}
