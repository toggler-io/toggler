package rollouts_test

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFeatureFlagChecker(t *testing.T) {
	t.Parallel()

	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	s.Parallel()

	s.Let(`PseudoRandPercentage`, func(v *testcase.V) interface{} { return int(0) })

	s.Before(func(t *testing.T, v *testcase.V) {
		require.Nil(t, GetStorage(v).Truncate(rollouts.FeatureFlag{}))
		require.Nil(t, GetStorage(v).Truncate(rollouts.Pilot{}))
	})

	s.Describe(`IsFeatureEnabledFor`, func(s *testcase.Spec) {
		SpecFeatureFlagChecker_IsFeatureEnabledFor(s)
	})

	s.Describe(`IsFeatureGloballyEnabled`, func(s *testcase.Spec) {
		SpecFeatureFlagChecker_IsFeatureGloballyEnabledFor(s)
	})
}

func SpecFeatureFlagChecker_IsFeatureGloballyEnabledFor(s *testcase.Spec) {
	subject := func(v *testcase.V) (bool, error) {
		return featureFlagChecker(v).IsFeatureGloballyEnabled(v.I(`FeatureName`).(string))
	}

	thenItWillReportThatFeatureNotGlobballyEnabled := func(s *testcase.Spec) {
		s.Then(`it will report that feature is not enabled globally`, func(t *testing.T, v *testcase.V) {
			enabled, err := subject(v)
			require.Nil(t, err)
			require.False(t, enabled)
		})
	}

	thenItWillReportThatFeatureIsGloballyRolledOut := func(s *testcase.Spec) {
		s.Then(`it will report that feature is globally rolled out`, func(t *testing.T, v *testcase.V) {
			enabled, err := subject(v)
			require.Nil(t, err)
			require.True(t, enabled)
		})
	}

	s.When(`feature flag is not seen before`, func(s *testcase.Spec) {
		s.Before(func(t *testing.T, v *testcase.V) {
			require.Nil(t, GetStorage(v).Truncate(rollouts.FeatureFlag{}))
		})

		thenItWillReportThatFeatureNotGlobballyEnabled(s)
	})

	s.When(`feature flag is given`, func(s *testcase.Spec) {
		s.Before(func(t *testing.T, v *testcase.V) {
			require.Nil(t, GetStorage(v).Save(GetFeatureFlag(v)))
		})

		s.And(`it is not yet rolled out globally`, func(s *testcase.Spec) {
			s.Let(`RolloutPercentage`, func(v *testcase.V) interface{} { return 99 })

			thenItWillReportThatFeatureNotGlobballyEnabled(s)
		})

		s.And(`it is rolled out globally`, func(s *testcase.Spec) {
			s.Let(`RolloutPercentage`, func(v *testcase.V) interface{} { return 100 })

			thenItWillReportThatFeatureIsGloballyRolledOut(s)
		})
	})
}

func SpecFeatureFlagChecker_IsFeatureEnabledFor(s *testcase.Spec) {
	subject := func(v *testcase.V) (bool, error) {
		return featureFlagChecker(v).IsFeatureEnabledFor(
			v.I(`FeatureName`).(string),
			v.I(`ExternalPilotID`).(string),
		)
	}

	s.When(`feature was never enabled before`, func(s *testcase.Spec) {
		s.Before(func(t *testing.T, v *testcase.V) {
			require.Nil(t, GetStorage(v).Truncate(rollouts.FeatureFlag{}))
		})

		s.Then(`it will tell that feature flag is not enabled`, func(t *testing.T, v *testcase.V) {
			ok, err := subject(v)
			require.Nil(t, err)
			require.False(t, ok)
		})
	})

	s.When(`feature is configured with rollout strategy`, func(s *testcase.Spec) {
		s.Before(func(t *testing.T, v *testcase.V) {
			require.Nil(t, GetStorage(v).Save(GetFeatureFlag(v)))
		})

		s.And(`by rollout percentage`, func(s *testcase.Spec) {
			s.And(`it is 0`, func(s *testcase.Spec) {
				s.Let(`RolloutPercentage`, func(v *testcase.V) interface{} { return int(0) })

				s.And(`pseudo rand percentage for the id`, func(s *testcase.Spec) {
					s.And(`it is 0 as well`, func(s *testcase.Spec) {
						s.Let(`PseudoRandPercentage`, func(v *testcase.V) interface{} { return int(0) })

						s.Then(`pilot is not enrolled for the feature`, func(t *testing.T, v *testcase.V) {
							ok, err := subject(v)
							require.Nil(t, err)
							require.False(t, ok)
						})
					})

					s.And(`it is greater than 0`, func(s *testcase.Spec) {
						s.Let(`PseudoRandPercentage`, func(v *testcase.V) interface{} { return int(42) })

						s.Then(`pilot is not enrolled for the feature`, func(t *testing.T, v *testcase.V) {
							ok, err := subject(v)
							require.Nil(t, err)
							require.False(t, ok)
						})
					})
				})
			})

			s.And(`it is greater than 0`, func(s *testcase.Spec) {
				s.Let(`RolloutPercentage`, func(v *testcase.V) interface{} { return rand.Intn(99) + 1 })

				s.And(`the next pseudo rand int is less or eq with the rollout percentage`, func(s *testcase.Spec) {
					s.Let(`PseudoRandPercentage`, func(v *testcase.V) interface{} {
						ffPerc := GetFeatureFlag(v).Rollout.Strategy.Percentage
						return ffPerc - rand.Intn(ffPerc)
					})

					s.Then(`it will enroll the pilot for the feature`, func(t *testing.T, v *testcase.V) {
						ok, err := subject(v)
						require.Nil(t, err)
						require.True(t, ok)
					})

					s.And(`the pilot is blacklisted manually from the feature`, func(s *testcase.Spec) {
						s.Before(func(t *testing.T, v *testcase.V) {

							pilot := &rollouts.Pilot{
								FeatureFlagID: GetFeatureFlag(v).ID,
								ExternalID:    v.I(`ExternalPilotID`).(string),
								Enrolled:      false,
							}

							require.Nil(t, GetStorage(v).Save(pilot))

						})

						s.Then(`the pilot will be not enrolled for the feature flag`, func(t *testing.T, v *testcase.V) {
							ok, err := subject(v)
							require.Nil(t, err)
							require.False(t, ok)
						})
					})

				})

				s.And(`the next pseudo rand int is is greater than the rollout percentage`, func(s *testcase.Spec) {

					s.Let(`PseudoRandPercentage`, func(v *testcase.V) interface{} {
						return GetFeatureFlag(v).Rollout.Strategy.Percentage + 1
					})

					s.Then(`pilot is not enrolled for the feature`, func(t *testing.T, v *testcase.V) {
						ok, err := subject(v)
						require.Nil(t, err)
						require.False(t, ok)
					})

					s.And(`if pilot manually enabled for the feature`, func(s *testcase.Spec) {
						s.Let(`PilotEnrollment`, func(v *testcase.V) interface{} { return true })

						s.Before(func(t *testing.T, v *testcase.V) {
							require.Nil(t, GetStorage(v).Save(GetPilot(v)))
						})

						s.Then(`the pilot is enrolled for the feature`, func(t *testing.T, v *testcase.V) {
							ok, err := subject(v)
							require.Nil(t, err)
							require.True(t, ok)
						})
					})
				})
			})

			s.And(`it is 100% or in other words it is set to be globally enabled`, func(s *testcase.Spec) {
				s.Let(`RolloutPercentage`, func(v *testcase.V) interface{} { return int(100) })

				s.And(`basically regardless the pseudo random percentage`, func(s *testcase.Spec) {
					s.Let(`PseudoRandPercentage`, func(v *testcase.V) interface{} { return rand.Intn(101) })

					s.Then(`it will always be enrolled`, func(t *testing.T, v *testcase.V) {
						ok, err := subject(v)
						require.Nil(t, err)
						require.True(t, ok)
					})
				})
			})

		})

		s.And(`by custom decision logic is defined with URL endpoint`, func(s *testcase.Spec) {
			s.Let(`RolloutApiURL`, func(v *testcase.V) interface{} {
				return v.I(`httptest.NewServer`).(*httptest.Server).URL
			})

			handler := func(t *testing.T, v *testcase.V) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, v.I(`FeatureName`).(string), r.URL.Query().Get(`feature-flag-name`))
					require.Equal(t, v.I(`ExternalPilotID`).(string), r.URL.Query().Get(`pilot-id`))
					w.WriteHeader(v.I(`replyCode`).(int))
				})
			}

			s.Let(`httptest.NewServer`, func(v *testcase.V) interface{} {
				return httptest.NewServer(handler(v.T(), v))
			})

			s.After(func(t *testing.T, v *testcase.V) {
				v.I(`httptest.NewServer`).(*httptest.Server).Close()
			})

			s.And(`the remote reject the pilot enrollment`, func(s *testcase.Spec) {
				s.Let(`replyCode`, func(v *testcase.V) interface{} { return rand.Intn(100) + 400 })

				s.Then(`the pilot is not enrolled for the feature`, func(t *testing.T, v *testcase.V) {
					enabled, err := subject(v)
					require.Nil(t, err)
					require.False(t, enabled)
				})
			})

			s.And(`the the remote accept the pilot enrollment`, func(s *testcase.Spec) {
				s.Let(`replyCode`, func(v *testcase.V) interface{} { return rand.Intn(100) + 200 })

				s.Then(`the pilot is not enrolled for the feature`, func(t *testing.T, v *testcase.V) {
					enabled, err := subject(v)
					require.Nil(t, err)
					require.True(t, enabled)
				})
			})
		})
	})
}

func featureFlagChecker(v *testcase.V) *rollouts.FeatureFlagChecker {
	return &rollouts.FeatureFlagChecker{
		Storage: v.I(`Storage`).(*Storage),
		IDPercentageCalculator: func(id string, seedSalt int64) (int, error) {
			require.Equal(v.T(), v.I(`ExternalPilotID`).(string), id)
			require.Equal(v.T(), v.I(`RolloutSeedSalt`).(int64), seedSalt)
			return v.I(`PseudoRandPercentage`).(int), nil
		},
	}
}
