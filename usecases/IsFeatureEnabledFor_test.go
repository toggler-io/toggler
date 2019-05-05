package usecases_test

import (
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"

	. "github.com/adamluzsi/FeatureFlags/testing"
)

func TestUseCases_IsFeatureEnabledFor(t *testing.T) {

	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	s.Parallel()

	UseCases := func(v *testcase.V) *usecases.UseCases { return v.I(`UseCases`).(*usecases.UseCases) }

	s.Let(`UseCases`, func(v *testcase.V) interface{} {
		return usecases.NewUseCases(v.I(`Storage`).(*Storage))
	})

	s.Describe(`IsFeatureEnabledFor`, func(s *testcase.Spec) {
		subject := func(v *testcase.V) (bool, error) {
			return UseCases(v).IsFeatureEnabledFor(
				v.I(`FeatureName`).(string),
				v.I(`ExternalPilotID`).(string),
			)
		}

		isEnrolled := func(t *testing.T, v *testcase.V) bool {
			enrolled, err := subject(v)
			require.Nil(t, err)
			return enrolled
		}

		s.When(`user piloting status registered`, func(s *testcase.Spec) {
			s.Before(func(t *testing.T, v *testcase.V) {
				enrollment := v.I(`enrollment`).(bool)
				require.Nil(t, UseCases(v).SetPilotEnrollmentForFeature(
					v.I(`FeatureName`).(string),
					v.I(`ExternalPilotID`).(string),
					enrollment))
			})

			s.And(`by whitelist`, func(s *testcase.Spec) {
				s.Let(`enrollment`, func(v *testcase.V) interface{} { return true })

				s.Then(`feature is enabled`, func(t *testing.T, v *testcase.V) {
					require.True(t, isEnrolled(t, v))
				})
			})

			s.And(`by blacklist`, func(s *testcase.Spec) {
				s.Let(`enrollment`, func(v *testcase.V) interface{} { return false })

				s.Then(`feature is disabled`, func(t *testing.T, v *testcase.V) {
					require.False(t, isEnrolled(t, v))
				})
			})
		})

		s.When(`many different user ask for feature enrollment`, func(s *testcase.Spec) {

			s.Let(`tolerationPercentage`, func(v *testcase.V) interface{} {
				var percentage int
				if testing.Short() {
					percentage = 5
				} else {
					percentage = 3
				}
				return percentage
			})

			s.Let(`samplingCount`, func(v *testcase.V) interface{} {
				var count int
				if testing.Short() {
					count = 1000
				} else {
					count = 10000
				}
				return count
			})

			s.Let(`extIDS`, func(v *testcase.V) interface{} {
				extIDS := []string{}
				samplingCount := v.I(`samplingCount`).(int)

				for i := 0; i < samplingCount; i++ {
					extIDS = append(extIDS, ExampleExternalPilotID())
				}

				return extIDS
			})

			s.And(`the rollout percentage is configured`, func(s *testcase.Spec) {

				s.Let(`expectedEnrollMaxPercentage`, func(v *testcase.V) interface{} {
					expectedEnrollMaxPercentage := rand.Intn(51) + 50
					tolerationPercentage := v.I(`tolerationPercentage`).(int)

					if 100 < expectedEnrollMaxPercentage+tolerationPercentage {
						tolerationPercentage = 100 - expectedEnrollMaxPercentage
					}

					return expectedEnrollMaxPercentage
				})

				s.Before(func(t *testing.T, v *testcase.V) {
					expectedEnrollMaxPercentage := v.I(`expectedEnrollMaxPercentage`).(int)

					require.Nil(t, UseCases(v).UpdateFeatureFlagRolloutPercentage(
						v.I(`FeatureName`).(string), expectedEnrollMaxPercentage))
				})

				s.Then(``, func(t *testing.T, v *testcase.V) {
					var enrolled, rejected int

					extIDS := v.I(`extIDS`).([]string)

					for _, extID := range extIDS {
						enrollment, err := UseCases(v).IsFeatureEnabledFor(
							v.I(`FeatureName`).(string), extID)

						require.Nil(t, err)

						if enrollment {
							enrolled++
						} else {
							rejected++
						}

					}

					expectedEnrollMaxPercentage := v.I(`expectedEnrollMaxPercentage`).(int)
					tolerationPercentage := v.I(`tolerationPercentage`).(int)

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
	})

}
