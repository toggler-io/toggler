package specs

import (
	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/toggler/testing"
	"testing"

	"github.com/adamluzsi/toggler/services/rollouts"

	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/stretchr/testify/require"
)

type FlagFinderSpec struct {
	Subject interface {
		rollouts.FlagFinder

		specs.MinimumRequirements
	}
}

func (spec FlagFinderSpec) Test(t *testing.T) {
	s := testcase.NewSpec(t)

	featureName := ExampleFeatureName()

	s.Describe(`FlagFinderSpec`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, spec.Subject.Truncate(rollouts.FeatureFlag{}))
		})

		s.After(func(t *testcase.T) {
			require.Nil(t, spec.Subject.Truncate(rollouts.FeatureFlag{}))
		})

		s.Describe(`FindFlagByName`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) *rollouts.FeatureFlag {
				ff, err := spec.Subject.FindFlagByName(featureName)
				require.Nil(t, err)
				return ff
			}

			s.When(`we don't have feature flag yet`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) { require.Nil(t, spec.Subject.Truncate(rollouts.FeatureFlag{})) })

				s.Then(`we receive back nil pointer`, func(t *testcase.T) {
					require.Nil(t, subject(t))
				})
			})

			s.When(`we have a feature flag already set`, func(s *testcase.Spec) {
				s.Let(`ff`, func(t *testcase.T) interface{} {
					return &rollouts.FeatureFlag{Name: featureName}
				})

				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.Save(t.I(`ff`).(*rollouts.FeatureFlag)))
				})

				s.Then(`searching for it returns the flag entity`, func(t *testcase.T) {
					ff := t.I(`ff`).(*rollouts.FeatureFlag)
					actually, err := spec.Subject.FindFlagByName(ff.Name)
					require.Nil(t, err)
					require.Equal(t, ff, actually)
				})
			})
		})
	})
}
