package usecases_test

import (
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testrun"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"

	. "github.com/adamluzsi/FeatureFlags/testing"
)

func TestUseCases_IsFeatureEnabledFor(t *testing.T) {
	t.Parallel()

	var (
		uc              *usecases.UseCases
		featureFlagName string
		externalPilotID string
		storage         usecases.Storage
	)

	subject := func() (bool, error) {
		return uc.IsFeatureEnabledFor(featureFlagName, externalPilotID)
	}

	isEnrolled := func(t *testing.T) bool {
		enrolled, err := subject()
		require.Nil(t, err)
		return enrolled
	}

	steps := testrun.Steps{}.Add(func(t *testing.T) {
		storage = NewStorage()
		uc = usecases.NewUseCases(storage)
		featureFlagName = ExampleFlagName()
		externalPilotID = ExampleExternalPilotID()
	})

	t.Run(`when user enrolled by white list`, func(t *testing.T) {
		steps := steps.Add(func(t *testing.T) {
			require.Nil(t, uc.SetPilotEnrollmentForFeature(featureFlagName, externalPilotID, true))
		})

		t.Run(`then feature is enabled`, func(t *testing.T) {
			steps.Setup(t)
			require.True(t, isEnrolled(t))
		})
	})

	t.Run(`when user blacklisted`, func(t *testing.T) {
		steps := steps.Add(func(t *testing.T) {
			require.Nil(t, uc.SetPilotEnrollmentForFeature(featureFlagName, externalPilotID, false))
		})

		t.Run(`then feature is enabled`, func(t *testing.T) {
			steps.Setup(t)
			require.False(t, isEnrolled(t))
		})
	})

	t.Run(`when many different user ask for feature enrollment`, func(t *testing.T) {
		var extIDS []string
		var tolerationPercentage int

		steps := steps.Add(func(t *testing.T) {

			tolerationPercentage = 3
			samplingCount := 10000

			if testing.Short() {
				tolerationPercentage = 5
				samplingCount = 1000
			}

			extIDS = []string{}
			for i := 0; i < samplingCount; i++ {
				extIDS = append(extIDS, ExampleExternalPilotID())
			}
		})

		t.Run(`and the rollout percentage is configured`, func(t *testing.T) {
			var expectedEnrollMaxPercentage int

			steps := steps.Add(func(t *testing.T) {
				expectedEnrollMaxPercentage = rand.Intn(51) + 50

				if 100 < expectedEnrollMaxPercentage + tolerationPercentage {
					tolerationPercentage = 100 - expectedEnrollMaxPercentage
				}

				require.Nil(t, uc.UpdateFeatureFlagRolloutPercentage(featureFlagName, expectedEnrollMaxPercentage))
			})

			t.Run(`then it is expected that the rollout percentage is honored somewhat`, func(t *testing.T) {
				steps.Setup(t)

				var enrolled, rejected int

				for _, extID := range extIDS {
					enrollment, err := uc.IsFeatureEnabledFor(featureFlagName, extID)

					require.Nil(t, err)

					if enrollment {
						enrolled++
					} else {
						rejected++
					}

				}

				t.Logf(`a little toleration is still accepted, as long in generally it is within the range (+%d%%)`, tolerationPercentage)
				maximumAcceptedEnrollmentPercentage := expectedEnrollMaxPercentage + tolerationPercentage
				minimumAcceptedEnrollmentPercentage := expectedEnrollMaxPercentage - tolerationPercentage

				t.Logf(`so the total percentage in this test that fulfil the requirements is %d%%`, maximumAcceptedEnrollmentPercentage)
				testRunResultPercentage := int(float32(enrolled) / float32(enrolled+rejected) * 100)

				t.Logf(`and the actual percentage is %d%%`, testRunResultPercentage)
				require.True(t, testRunResultPercentage <= maximumAcceptedEnrollmentPercentage)
				require.True(t, minimumAcceptedEnrollmentPercentage <= testRunResultPercentage)

			})
		})

	})
}
