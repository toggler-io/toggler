package specs

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/stretchr/testify/require"
	"testing"
)

type FlagFinderSpec struct {
	Subject interface {
		rollouts.FlagFinder

		specs.MinimumRequirements
	}
}

func (spec FlagFinderSpec) Test(t *testing.T) {
	featureName := exampleFeatureFlagName()

	setup := func(t *testing.T) {
		require.Nil(t, spec.Subject.Truncate(rollouts.FeatureFlag{}))
	}

	t.Run(`FindByFlagName`, func(t *testing.T) {

		subject := func(t *testing.T) *rollouts.FeatureFlag {
			ff, err := spec.Subject.FindByFlagName(featureName)
			require.Nil(t, err)
			return ff
		}

		t.Run(`when we don't have feature flag yet`, func(t *testing.T) {
			setup(t)

			t.Run(`then we receive back nil pointer`, func(t *testing.T) {
				require.Nil(t, subject(t))
			})
		})

		t.Run(`when we have a feature flag already set`, func(t *testing.T) {
			setup(t)

			ff := &rollouts.FeatureFlag{Name: featureName}
			require.Nil(t, spec.Subject.Save(ff))

			t.Run(`then searching for it returns the flag entity`, func(t *testing.T) {
				actually, err := spec.Subject.FindByFlagName(ff.Name)
				require.Nil(t, err)
				require.Equal(t, ff, &actually)
			})

		})

	})
}
