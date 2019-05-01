package rollouts_test

import (
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFeatureFlagChecker(t *testing.T) {
	t.Parallel()

	ExternalPilotID := ExampleExternalPilotID()
	FeatureFlagName := ExampleFlagName()
	RolloutSeedSalt := time.Now().Unix()
	storage := NewStorage()

	var PseudoRandPercentage int

	featureFlagChecker := func() *rollouts.FeatureFlagChecker {
		return &rollouts.FeatureFlagChecker{
			Storage: storage,
			IDPercentageCalculator: func(id string, seedSalt int64) (int, error) {
				require.Equal(t, ExternalPilotID, id)
				require.Equal(t, RolloutSeedSalt, seedSalt)
				return PseudoRandPercentage, nil
			},
		}
	}

	var ff *rollouts.FeatureFlag

	setup := func(t *testing.T, ffSetup func(*rollouts.FeatureFlag)) {
		require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))
		require.Nil(t, storage.Truncate(rollouts.Pilot{}))

		ff = &rollouts.FeatureFlag{Name: FeatureFlagName}
		ff.Rollout.RandSeedSalt = RolloutSeedSalt

		require.Nil(t, storage.Save(ff))

		if ffSetup != nil {
			ffSetup(ff)
		}
	}

	t.Run(`IsFeatureEnabledFor`, func(t *testing.T) {
		subject := func() (bool, error) {
			return featureFlagChecker().IsFeatureEnabledFor(FeatureFlagName, ExternalPilotID)
		}

		t.Run(`when feature was never enabled before`, func(t *testing.T) {
			require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))

			t.Run(`then it will tell that feature flag is not enabled`, func(t *testing.T) {
				ok, err := subject()
				require.Nil(t, err)
				require.False(t, ok)
			})
		})
		t.Run(`when feature is configured for rollout`, func(t *testing.T) {
			t.Run(`and the rollout percentage`, func(t *testing.T) {
				t.Run(`is 0`, func(t *testing.T) {
					setup(t, func(flag *rollouts.FeatureFlag) {
						flag.Rollout.Strategy.Percentage = 0
					})

					t.Run(`and pseudo rand percentage for the id is 0 as well`, func(t *testing.T) {
						PseudoRandPercentage = 0

						t.Run(`then pilot is not enrolled for the feature`, func(t *testing.T) {
							ok, err := subject()
							require.Nil(t, err)
							require.False(t, ok)
						})
					})

					t.Run(`is greater than 0`, func(t *testing.T) {
						PseudoRandPercentage = 42

						t.Run(`then pilot is not enrolled for the feature`, func(t *testing.T) {
							ok, err := subject()
							require.Nil(t, err)
							require.False(t, ok)
						})
					})
				})
				t.Run(`is greater than 0`, func(t *testing.T) {
					setRollout := func(flag *rollouts.FeatureFlag) {
						flag.Rollout.Strategy.Percentage = rand.Intn(99) + 1
					}

					t.Run(`and the next pseudo rand int`, func(t *testing.T) {
						defer func() { PseudoRandPercentage = 0 }()

						t.Run(`is less or eq with the rollout percentage`, func(t *testing.T) {
							setup(t, setRollout)
							PseudoRandPercentage = ff.Rollout.Strategy.Percentage - rand.Intn(ff.Rollout.Strategy.Percentage)

							t.Run(`then it will enroll the pilot for the feature`, func(t *testing.T) {
								ok, err := subject()
								require.Nil(t, err)
								require.True(t, ok)
							})

							t.Run(`and the pilot is blacklisted manually from the feature`, func(t *testing.T) {
								require.Nil(t, storage.Save(&rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPilotID, Enrolled: false}))

								t.Run(`the the pilot will be not enrolled for the feature flag`, func(t *testing.T) {
									ok, err := subject()
									require.Nil(t, err)
									require.False(t, ok)
								})
							})
						})

						t.Run(`is greater than the rollout percentage`, func(t *testing.T) {
							setup(t, setRollout)
							PseudoRandPercentage = ff.Rollout.Strategy.Percentage + 1
							defer func() { PseudoRandPercentage = 0 }()

							t.Run(`then pilot is not enrolled for the feature`, func(t *testing.T) {
								ok, err := subject()
								require.Nil(t, err)
								require.False(t, ok)

								t.Run(`but when pilot manually enabled for the feature`, func(t *testing.T) {
									require.Nil(t, storage.Save(&rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPilotID, Enrolled: true}))

									t.Run(`then the pilot is enrolled for the feature`, func(t *testing.T) {
										ok, err := subject()
										require.Nil(t, err)
										require.True(t, ok)
									})
								})
							})
						})
					})

				})
				t.Run(`is 100, in other words it is set to be globally enabled`, func(t *testing.T) {
					setup(t, func(flag *rollouts.FeatureFlag) {
						ff.Rollout.Strategy.Percentage = 100
					})
					t.Run(`and basically regardless the pseudo random percentage`, func(t *testing.T) {
						PseudoRandPercentage = rand.Intn(101)
						defer func() { PseudoRandPercentage = 0 }()

						t.Run(`then it will always be enrolled`, func(t *testing.T) {
							ok, err := subject()
							require.Nil(t, err)
							require.True(t, ok)
						})
					})
				})
			})
			t.Run(`and custom decision logic is defined with URL endpoint`, func(t *testing.T) {
				var url string
				flagSetup := func(flag *rollouts.FeatureFlag) {
					flag.Rollout.Strategy.URL = url
				}

				var replyCode int
				handler := func(t *testing.T) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						require.Equal(t, FeatureFlagName, r.URL.Query().Get(`feature-flag-name`))
						require.Equal(t, ExternalPilotID, r.URL.Query().Get(`pilot-id`))
						w.WriteHeader(replyCode)
					})
				}

				t.Run(`and the remote reject the pilot enrollment`, func(t *testing.T) {
					replyCode = rand.Intn(100) + 400

					t.Run(`then the pilot is not enrolled for the feature`, func(t *testing.T) {
						s := httptest.NewServer(handler(t))
						defer s.Close()
						url = s.URL
						setup(t, flagSetup)

						enabled, err := subject()
						require.Nil(t, err)
						require.False(t, enabled)
					})
				})

				t.Run(`and the the remote accept the pilot enrollment`, func(t *testing.T) {
					replyCode = rand.Intn(100) + 200

					t.Run(`then the pilot is not enrolled for the feature`, func(t *testing.T) {
						s := httptest.NewServer(handler(t))
						defer s.Close()
						url = s.URL
						setup(t, flagSetup)

						enabled, err := subject()
						require.Nil(t, err)
						require.True(t, enabled)
					})
				})


			})
		})
	})

	t.Run(`IsFeatureGloballyEnabled`, func(t *testing.T) {
		subject := func() (bool, error) {
			return featureFlagChecker().IsFeatureGloballyEnabled(FeatureFlagName)
		}

		thenItWillReportThatFeatureNotGlobballyEnabled := func(t *testing.T) {
			t.Run(`then it will report that feature is not enabled globally`, func(t *testing.T) {
				enabled, err := subject()
				require.Nil(t, err)
				require.False(t, enabled)
			})
		}

		thenItWillReportThatFeatureIsGloballyRolledOut := func(t *testing.T) {
			t.Run(`then it will report that feature is globally rolled out`, func(t *testing.T) {
				enabled, err := subject()
				require.Nil(t, err)
				require.True(t, enabled)
			})
		}

		t.Run(`when feature flag is not seen before`, func(t *testing.T) {
			setup(t, func(flag *rollouts.FeatureFlag) {
				require.Nil(t, storage.DeleteByID(flag, flag.ID))
			})

			thenItWillReportThatFeatureNotGlobballyEnabled(t)
		})

		t.Run(`when feature flag is given`, func(t *testing.T) {
			t.Run(`and it is not yet rolled out globally`, func(t *testing.T) {
				setup(t, func(flag *rollouts.FeatureFlag) {
					flag.Rollout.Strategy.Percentage = 99
				})

				thenItWillReportThatFeatureNotGlobballyEnabled(t)
			})

			t.Run(`and it is rolled out globally`, func(t *testing.T) {
				setup(t, func(flag *rollouts.FeatureFlag) {
					ff.Rollout.Strategy.Percentage = 100
				})

				thenItWillReportThatFeatureIsGloballyRolledOut(t)
			})
		})

	})
}
