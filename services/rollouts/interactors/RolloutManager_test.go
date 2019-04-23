package interactors_test

import (
	"math/rand"
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
	thelp "github.com/adamluzsi/FeatureFlags/services/rollouts/testing"
	"github.com/stretchr/testify/require"
)

func TestRolloutTrier(t *testing.T) {
	t.Parallel()

	ExternalPilotID := randomdata.MacAddress()
	flagName := randomdata.SillyName()
	var ff *rollouts.FeatureFlag

	var nextRandIntn int
	storage := thelp.NewStorage()

	trier := func() *interactors.RolloutManager {
		return &interactors.RolloutManager{
			Storage: storage,
			RandIntn: func(max int) int {
				return nextRandIntn
			},
		}
	}

	setup := func(t *testing.T, ffSetup func(*rollouts.FeatureFlag)) {
		require.Nil(t, storage.Truncate(rollouts.FeatureFlag{}))
		require.Nil(t, storage.Truncate(rollouts.Pilot{}))

		ff = &rollouts.FeatureFlag{Name: flagName}

		if ffSetup != nil {
			ffSetup(ff)
		}

		require.Nil(t, storage.Save(ff))
	}

	pilotEnrollment := func(t *testing.T) bool {
		require.NotNil(t, ff, `to use this, it is expected that the feature flag already exist`)
		pilot, err := storage.FindFlagPilotByExternalPilotID(ff.ID, ExternalPilotID)

		require.Nil(t, err)
		if pilot == nil {
			return false
		}

		return pilot.Enrolled
	}

	nextRandValueIs := func(NextValue int) func() {
		nextRandIntn = NextValue
		return func() { nextRandIntn = 0 }
	}

	t.Run(`TryRolloutThisPilot`, func(t *testing.T) {
		subject := func() error {
			return trier().TryRolloutThisPilot(flagName, ExternalPilotID)
		}

		t.Run(`when rollout percentage`, func(t *testing.T) {
			t.Run(`is 0`, func(t *testing.T) {
				setup(t, nil)

				t.Run(`and the next pseudo rand int`, func(t *testing.T) {
					t.Run(`is 0 as well`, func(t *testing.T) {
						defer nextRandValueIs(0)()

						t.Run(`then pilot is not enrolled for the feature`, func(t *testing.T) {
							require.Nil(t, subject())

							require.False(t, pilotEnrollment(t))
						})
					})

					t.Run(`is greater than 0`, func(t *testing.T) {
						defer nextRandValueIs(42)()

						t.Run(`then pilot is not enrolled for the feature`, func(t *testing.T) {
							require.Nil(t, subject())

							require.False(t, pilotEnrollment(t))
						})
					})
				})
			})

			t.Run(`is greater than 0`, func(t *testing.T) {
				setRollout := func(flag *rollouts.FeatureFlag) {
					flag.Rollout.Percentage = rand.Intn(99) + 1
				}

				t.Run(`and the next pseudo rand int`, func(t *testing.T) {

					t.Run(`is less or eq with the rollout percentage`, func(t *testing.T) {
						setup(t, setRollout)
						defer nextRandValueIs(ff.Rollout.Percentage - rand.Intn(ff.Rollout.Percentage))()

						t.Run(`then it will enroll the pilot for the feature`, func(t *testing.T) {
							require.Nil(t, subject())
							require.True(t, pilotEnrollment(t))

							p, err := storage.FindFlagPilotByExternalPilotID(ff.ID, ExternalPilotID)
							require.Nil(t, err)
							require.NotNil(t, p)
						})
					})

					t.Run(`is greater than the rollout percentage`, func(t *testing.T) {
						setup(t, setRollout)
						defer nextRandValueIs(ff.Rollout.Percentage + 1)()

						t.Run(`then pilot is not enrolled for the feature`, func(t *testing.T) {
							require.Nil(t, subject())

							require.False(t, pilotEnrollment(t))

							t.Run(`and after the first call`, func(t *testing.T) {
								t.Run(`when rand int less than accepted percentage`, func(t *testing.T) {
									defer nextRandValueIs(ff.Rollout.Percentage - rand.Intn(ff.Rollout.Percentage))()

									t.Run(`then already seen user will be still blacklisted from being enrolled for this feature`, func(t *testing.T) {
										require.Nil(t, subject())

										require.False(t, pilotEnrollment(t))
									})
								})
							})
						})
					})
				})

			})
		})

	})
}
