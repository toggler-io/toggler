package specs

import (
	"context"
	context2 "context"
	"strconv"
	"testing"

	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/toggler/testing"

	"github.com/adamluzsi/toggler/services/rollouts"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/stretchr/testify/require"
)

type PilotFinderSpec struct {
	Subject interface {
		rollouts.PilotFinder

		specs.MinimumRequirements
	}
}

func (spec PilotFinderSpec) Test(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`PilotFinderSpec`, func(s *testcase.Spec) {

		s.Let(`flagName`, func(t *testcase.T) interface{} {
			return ExampleFeatureName()
		})

		s.Before(func(t *testcase.T) {
			require.Nil(t, spec.Subject.Truncate(context.Background(), rollouts.FeatureFlag{}))
			require.Nil(t, spec.Subject.Truncate(context.Background(), rollouts.Pilot{}))
		})

		s.After(func(t *testcase.T) {
			require.Nil(t, spec.Subject.Truncate(context.Background(), rollouts.FeatureFlag{}))
			require.Nil(t, spec.Subject.Truncate(context.Background(), rollouts.Pilot{}))
		})

		s.Describe(`FindPilotsByFeatureFlag`, func(s *testcase.Spec) {
			getFF := func(t *testcase.T) *rollouts.FeatureFlag {
				var ff *rollouts.FeatureFlag
				f := t.I(`ff`)
				if f != nil {
					ff = f.(*rollouts.FeatureFlag)
				}
				return ff
			}

			subject := func(t *testcase.T) frameless.Iterator {
				return spec.Subject.FindPilotsByFeatureFlag(context2.TODO(), getFF(t))
			}

			thenNoPilotsFound := func(s *testcase.Spec) {
				s.Then(`no pilots found`, func(t *testcase.T) {
					iter := subject(t)
					require.NotNil(t, iter)
					require.False(t, iter.Next())
					require.Nil(t, iter.Err())
				})
			}

			s.When(`feature object is nil`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.Truncate(context.Background(), rollouts.FeatureFlag{}))
				})
				s.Let(`ff`, func(t *testcase.T) interface{} { return nil })
				thenNoPilotsFound(s)
			})

			s.When(`feature object has no reference`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.Truncate(context.Background(), rollouts.FeatureFlag{}))
				})
				s.Let(`ff`, func(t *testcase.T) interface{} { return &rollouts.FeatureFlag{} })
				thenNoPilotsFound(s)
			})

			s.When(`feature flag exists`, func(s *testcase.Spec) {
				s.Let(`ff`, func(t *testcase.T) interface{} {
					ff := &rollouts.FeatureFlag{Name: t.I(`flagName`).(string)}
					require.Nil(t, spec.Subject.Save(context.Background(),ff))
					return ff
				})

				thenNoPilotsFound(s)

				s.And(`there are registered pilots for the feature`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						expectedPilots := t.I(`expectedPilots`).([]*rollouts.Pilot)

						for _, pilot := range expectedPilots {
							require.Nil(t, spec.Subject.Save(context.Background(),pilot))
						}
					})

					s.Let(`expectedPilots`, func(t *testcase.T) interface{} {
						var expectedPilots []*rollouts.Pilot
						ff := t.I(`ff`).(*rollouts.FeatureFlag)

						for i := 0; i < 5; i++ {
							pilot := &rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: strconv.Itoa(i)}
							expectedPilots = append(expectedPilots, pilot)
						}

						return expectedPilots
					})

					s.Then(`it will return all of them`, func(t *testcase.T) {
						iter := subject(t)
						defer iter.Close()
						require.NotNil(t, iter)

						var actualPilots []*rollouts.Pilot

						for iter.Next() {
							var actually rollouts.Pilot
							require.Nil(t, iter.Decode(&actually))
							actualPilots = append(actualPilots, &actually)
						}

						require.Nil(t, iter.Err())

						expectedPilots := t.I(`expectedPilots`).([]*rollouts.Pilot)

						require.True(t, len(expectedPilots) == len(actualPilots))

						for _, expected := range expectedPilots {
							require.Contains(t, actualPilots, expected)
						}
					})
				})
			})
		})

		s.Describe(`FindFlagPilotByExternalPilotID`, func(s *testcase.Spec) {
			const ExternalPublicPilotID = `42`

			subject := func(t *testcase.T) (*rollouts.Pilot, error) {
				return spec.Subject.FindFlagPilotByExternalPilotID(context2.TODO(), t.I(`featureFlagID`).(string), ExternalPublicPilotID)
			}

			ThenNoPilotsFound := func(s *testcase.Spec) {
				s.Then(`no pilots found`, func(t *testcase.T) {
					pilot, err := subject(t)
					require.Nil(t, err)
					require.Nil(t, pilot)
				})
			}

			s.When(`feature was never enabled before`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.Truncate(context.Background(), rollouts.FeatureFlag{}))
				})
				s.Let(`featureFlagID`, func(t *testcase.T) interface{} { return "not exinsting ID" })
				ThenNoPilotsFound(s)
			})

			s.When(`feature flag exists`, func(s *testcase.Spec) {
				s.Let(`featureFlagID`, func(t *testcase.T) interface{} {
					ff := &rollouts.FeatureFlag{Name: t.I(`flagName`).(string)}
					ff.Rollout.Strategy.Percentage = 100
					require.Nil(t, spec.Subject.Save(context.Background(),ff))
					return ff.ID
				})

				ThenNoPilotsFound(s)

				s.And(`the given there is a registered pilot for the feature`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						require.Nil(t, spec.Subject.Truncate(context.Background(), rollouts.Pilot{}))
						featureFlagID := t.I(`featureFlagID`).(string)
						pilot := &rollouts.Pilot{FeatureFlagID: featureFlagID, ExternalID: ExternalPublicPilotID}
						require.Nil(t, spec.Subject.Save(context.Background(),pilot))
					})

					s.Then(`asd`, func(t *testcase.T) {
						pilot, err := subject(t)
						require.Nil(t, err)
						require.NotNil(t, pilot)

						featureFlagID := t.I(`featureFlagID`).(string)
						require.Equal(t, ExternalPublicPilotID, pilot.ExternalID)
						require.Equal(t, featureFlagID, pilot.FeatureFlagID)
					})
				})
			})
		})
	})
}
