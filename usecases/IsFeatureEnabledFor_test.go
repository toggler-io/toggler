package usecases_test

import (
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"

	. "github.com/adamluzsi/FeatureFlags/testing"
)

func TestUseCases_IsFeatureEnabledFor(t *testing.T) {

	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	SetupSpec(s)
	s.Parallel()

	s.Describe(`IsFeatureEnabledFor`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (bool, error) {
			return GetUseCases(t).IsFeatureEnabledFor(
				t.I(`FeatureName`).(string),
				t.I(`ExternalPilotID`).(string),
			)
		}

		isEnrolled := func(t *testcase.T) bool {
			enrolled, err := subject(t)
			require.Nil(t, err)
			return enrolled
		}

		s.When(`user piloting status registered`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				enrollment := t.I(`enrollment`).(bool)

				require.Nil(t, GetRolloutManager(t).SetPilotEnrollmentForFeature(
					t.I(`FeatureName`).(string),
					t.I(`ExternalPilotID`).(string),
					enrollment,
				))
			})

			s.And(`by whitelist`, func(s *testcase.Spec) {
				s.Let(`enrollment`, func(t *testcase.T) interface{} { return true })

				s.Then(`feature is enabled`, func(t *testcase.T) {
					require.True(t, isEnrolled(t))
				})
			})

			s.And(`by blacklist`, func(s *testcase.Spec) {
				s.Let(`enrollment`, func(t *testcase.T) interface{} { return false })

				s.Then(`feature is disabled`, func(t *testcase.T) {
					require.False(t, isEnrolled(t))
				})
			})
		})

		s.When(`many different user ask for feature enrollment`, func(s *testcase.Spec) {

			s.Let(`tolerationPercentage`, func(t *testcase.T) interface{} {
				var percentage int
				if testing.Short() {
					percentage = 5
				} else {
					percentage = 3
				}
				return percentage
			})

			s.Let(`samplingCount`, func(t *testcase.T) interface{} {
				var count int
				if testing.Short() {
					count = 1000
				} else {
					count = 10000
				}
				return count
			})

			s.Let(`extIDS`, func(t *testcase.T) interface{} {
				extIDS := []string{}
				samplingCount := t.I(`samplingCount`).(int)

				for i := 0; i < samplingCount; i++ {
					extIDS = append(extIDS, ExampleExternalPilotID())
				}

				return extIDS
			})

			s.And(`the rollout percentage is configured`, func(s *testcase.Spec) {

				s.Let(`expectedEnrollMaxPercentage`, func(t *testcase.T) interface{} {
					expectedEnrollMaxPercentage := rand.Intn(51) + 50
					tolerationPercentage := t.I(`tolerationPercentage`).(int)

					if 100 < expectedEnrollMaxPercentage+tolerationPercentage {
						tolerationPercentage = 100 - expectedEnrollMaxPercentage
					}

					return expectedEnrollMaxPercentage
				})

				s.Before(func(t *testcase.T) {
					expectedEnrollMaxPercentage := t.I(`expectedEnrollMaxPercentage`).(int)

					require.Nil(t, GetUseCases(t).UpdateFeatureFlagRolloutPercentage(
						t.I(`FeatureName`).(string), expectedEnrollMaxPercentage))
				})

				s.Then(``, func(t *testcase.T) {
					var enrolled, rejected int

					extIDS := t.I(`extIDS`).([]string)

					for _, extID := range extIDS {
						enrollment, err := GetUseCases(t).IsFeatureEnabledFor(
							t.I(`FeatureName`).(string), extID)

						require.Nil(t, err)

						if enrollment {
							enrolled++
						} else {
							rejected++
						}

					}

					expectedEnrollMaxPercentage := t.I(`expectedEnrollMaxPercentage`).(int)
					tolerationPercentage := t.I(`tolerationPercentage`).(int)

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
