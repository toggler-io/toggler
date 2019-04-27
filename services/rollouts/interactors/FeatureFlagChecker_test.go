package interactors_test

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
	. "github.com/adamluzsi/FeatureFlags/services/rollouts/testing"
	"github.com/stretchr/testify/require"
)

func TestFeatureFlagChecker(t *testing.T) {
	t.Parallel()

	ExternalPilotID := ExampleExternalPilotID()
	FeatureFlagName := ExampleFlagName()
	storage := NewStorage()

	var PseudoRandPercentage int

	featureFlagChecker := func() *interactors.FeatureFlagChecker {
		return &interactors.FeatureFlagChecker{
			Storage:                storage,
			IDPercentageCalculator: func(string) int { return PseudoRandPercentage },
		}
	}

	var ff *rollouts.FeatureFlag

	setup := func(t *testing.T, ffSetup func(*rollouts.FeatureFlag)) {
		require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))
		require.Nil(t, storage.Truncate(rollouts.Pilot{}))

		ff = &rollouts.FeatureFlag{Name: FeatureFlagName}

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
						flag.Rollout.Percentage = 0
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
						flag.Rollout.Percentage = rand.Intn(99) + 1
					}

					t.Run(`and the next pseudo rand int`, func(t *testing.T) {
						defer func() { PseudoRandPercentage = 0 }()

						t.Run(`is less or eq with the rollout percentage`, func(t *testing.T) {
							setup(t, setRollout)
							PseudoRandPercentage = ff.Rollout.Percentage - rand.Intn(ff.Rollout.Percentage)

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
							PseudoRandPercentage = ff.Rollout.Percentage + 1
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
						ff.Rollout.Percentage = 100
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
					flag.Rollout.Percentage = 99
				})

				thenItWillReportThatFeatureNotGlobballyEnabled(t)
			})

			t.Run(`and it is rolled out globally`, func(t *testing.T) {
				setup(t, func(flag *rollouts.FeatureFlag) {
					ff.Rollout.Percentage = 100
				})

				thenItWillReportThatFeatureIsGloballyRolledOut(t)
			})
		})

	})
}

func TestPseudoRandPercentageWithFNV1a64(t *testing.T) {
	subject := interactors.PseudoRandPercentageWithFNV1a64

	t.Run(`it is expected that the result is deterministic`, func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 1000; i++ {
			res1 := subject(strconv.Itoa(i))
			res2 := subject(strconv.Itoa(i))
			require.Equal(t, res1, res2)
		}
	})

	t.Run(`it is expected that the values are between 0 and 100`, func(t *testing.T) {
		t.Parallel()

		var minFount, maxFount bool

		for i := 0; i<=10000; i++ {
			res := subject(strconv.Itoa(i))

			require.True(t, 0 <= res && res <= 100)

			if res == 0 {
				minFount = true
			}

			if res == 100 {
				maxFount = true
			}
		}

		require.True(t, minFount)
		require.True(t, maxFount)
	})
}
