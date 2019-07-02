package rollouts_test

import (
	"context"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/toggler/services/rollouts"
	. "github.com/adamluzsi/toggler/testing"
	"github.com/stretchr/testify/require"
)

func TestFeatureFlagChecker(t *testing.T) {
	t.Parallel()

	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	s.Parallel()

	s.Let(`PseudoRandPercentage`, func(t *testcase.T) interface{} { return int(0) })

	s.Before(func(t *testcase.T) {
		require.Nil(t, GetStorage(t).Truncate(context.Background(), rollouts.FeatureFlag{}))
		require.Nil(t, GetStorage(t).Truncate(context.Background(), rollouts.Pilot{}))
	})

	s.Describe(`IsFeatureEnabledFor`, func(s *testcase.Spec) {
		SpecFeatureFlagChecker_IsFeatureEnabledFor(s)
	})

	s.Describe(`IsFeatureGloballyEnabled`, func(s *testcase.Spec) {
		SpecFeatureFlagChecker_IsFeatureGloballyEnabledFor(s)
	})
}

func SpecFeatureFlagChecker_IsFeatureGloballyEnabledFor(s *testcase.Spec) {
	subject := func(t *testcase.T) (bool, error) {
		return featureFlagChecker(t).IsFeatureGloballyEnabled(t.I(`FeatureName`).(string))
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
			require.Nil(t, GetStorage(t).Truncate(context.Background(), rollouts.FeatureFlag{}))
		})

		thenItWillReportThatFeatureNotGlobballyEnabled(s)
	})

	s.When(`feature flag is given`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Save(context.TODO(), GetFeatureFlag(t)))
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
			t.I(`FeatureName`).(string),
			t.I(`ExternalPilotID`).(string),
		)
	}

	s.When(`feature was never enabled before`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Truncate(context.Background(), rollouts.FeatureFlag{}))
		})

		s.Then(`it will tell that feature flag is not enabled`, func(t *testcase.T) {
			ok, err := subject(t)
			require.Nil(t, err)
			require.False(t, ok)
		})
	})

	s.When(`feature is configured with rollout strategy`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Save(context.TODO(), GetFeatureFlag(t)))
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
						ffPerc := GetFeatureFlag(t).Rollout.Strategy.Percentage
						return ffPerc - rand.Intn(ffPerc)
					})

					s.Then(`it will enroll the pilot for the feature`, func(t *testcase.T) {
						ok, err := subject(t)
						require.Nil(t, err)
						require.True(t, ok)
					})

					s.And(`the pilot is blacklisted manually from the feature`, func(s *testcase.Spec) {
						s.Before(func(t *testcase.T) {

							pilot := &rollouts.Pilot{
								FeatureFlagID: GetFeatureFlag(t).ID,
								ExternalID:    t.I(`ExternalPilotID`).(string),
								Enrolled:      false,
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
						return GetFeatureFlag(t).Rollout.Strategy.Percentage + 1
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
					require.Equal(t, t.I(`FeatureName`).(string), r.URL.Query().Get(`feature`))
					require.Equal(t, t.I(`ExternalPilotID`).(string), r.URL.Query().Get(`id`))
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

func featureFlagChecker(t *testcase.T) *rollouts.FeatureFlagChecker {
	return &rollouts.FeatureFlagChecker{
		Storage: t.I(`TestStorage`).(*TestStorage),
		IDPercentageCalculator: func(id string, seedSalt int64) (int, error) {
			require.Equal(t, t.I(`ExternalPilotID`).(string), id)
			require.Equal(t, t.I(`RolloutSeedSalt`).(int64), seedSalt)
			return t.I(`PseudoRandPercentage`).(int), nil
		},
	}
}
