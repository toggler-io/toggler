package release_test

import (
	"context"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/services/release"
	. "github.com/toggler-io/toggler/testing"
	"github.com/stretchr/testify/require"
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

	s.Describe(`IsFeatureEnabledFor`, func(s *testcase.Spec) {
		SpecFeatureFlagChecker_IsFeatureEnabledFor(s)
	})

	s.Describe(`IsFeatureGloballyEnabled`, func(s *testcase.Spec) {
		SpecFeatureFlagChecker_IsFeatureGloballyEnabledFor(s)
	})

	s.Describe(`GetPilotFlagStates`, func(s *testcase.Spec) {
		SpecFeatureFlagChecker_GetPilotFlagStates(s)
	})
}

func SpecFeatureFlagChecker_GetPilotFlagStates(s *testcase.Spec) {
	subject := func(t *testcase.T) (map[string]bool, error) {
		ctx := context.Background()
		return featureFlagChecker(t).GetPilotFlagStates(ctx,
			GetExternalPilotID(t), t.I(`flag names`).([]string)...)
	}

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

	s.Let(`flag names`, func(t *testcase.T) interface{} {
		return []string{GetReleaseFlagName(t), `non-existent`}
	})

	s.Before(cleanup)
	s.After(cleanup)

	s.When(`no feature flag is set`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Truncate(context.Background(), release.Flag{}))
		})

		s.Then(`it will return with flipped off states`, func(t *testcase.T) {
			states, err := subject(t)
			require.Nil(t, err)
			t.Log(states)
			require.Equal(t, 2, len(states))
			stateIs(t, states, GetReleaseFlagName(t), false)
			stateIs(t, states, `non-existent`, false)
		})
	})

	s.When(`some of the feature flag exists`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Save(context.Background(), GetReleaseFlag(t)))
		})

		// fix the value, so the dependency sub context can work with an assumption
		s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return int(42) })

		s.And(`the pilot win the enrollment dice roll`, func(s *testcase.Spec) {
			s.Let(`PseudoRandPercentage`, func(t *testcase.T) interface{} { return int(0) })

			s.Then(`it is expected to receive switched ON state for the flag`, func(t *testcase.T) {
				states, err := subject(t)
				require.Nil(t, err)
				stateIs(t, states, GetReleaseFlagName(t), true)
			})

			s.And(`the pilot is black listed`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, GetStorage(t).Save(context.Background(), &release.Pilot{
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

		s.And(`the pilot lose the enrollment dice roll`, func(s *testcase.Spec) {
			s.Let(`PseudoRandPercentage`, func(t *testcase.T) interface{} { return int(100) })

			s.Then(`it is expected to receive switched OFF state for the flag`, func(t *testcase.T) {
				states, err := subject(t)
				require.Nil(t, err)
				stateIs(t, states, GetReleaseFlagName(t), false)
			})

			s.And(`the pilot is white listed`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, GetStorage(t).Save(context.Background(), &release.Pilot{
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
		})

		s.Then(`it will return with flipped states`, func(t *testcase.T) {
			states, err := subject(t)
			require.Nil(t, err)
			require.Equal(t, 2, len(states))

			for _, flagName := range t.I(`flag names`).([]string) {
				if _, ok := states[flagName]; !ok {
					t.Fatalf(`feature flag not included in the return states: %s`, flagName)
				}
			}
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
			require.Nil(t, GetStorage(t).Save(context.TODO(), GetReleaseFlag(t)))
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

func SpecFeatureFlagChecker_IsFeatureEnabledFor(s *testcase.Spec) {
	subject := func(t *testcase.T) (bool, error) {
		return featureFlagChecker(t).IsFeatureEnabledFor(
			t.I(`ReleaseFlagName`).(string),
			t.I(`PilotExternalID`).(string),
		)
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
			require.Nil(t, GetStorage(t).Save(context.TODO(), GetReleaseFlag(t)))
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

							require.Nil(t, GetStorage(t).Save(context.TODO(), pilot))

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
							require.Nil(t, GetStorage(t).Save(context.TODO(), GetPilot(t)))
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
