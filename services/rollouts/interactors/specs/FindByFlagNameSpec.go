package specs

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/stretchr/testify/require"
	"testing"
)

type FindByFlagNameSpec struct {
	Subject interface {
		interactors.FindByFlagName

		specs.MinimumRequirements
		specs.Purge
	}
}

func (spec FindByFlagNameSpec) Test(t *testing.T) {
	featureName := "feature-name"

	setup := func(t *testing.T) {
		require.Nil(t, spec.Subject.Truncate(rollouts.FeatureFlag{}))
	}

	t.Run(`given we don't have feature flag yet`, func(t *testing.T) {
		setup(t)

		var actually rollouts.FeatureFlag

		ok, err := spec.Subject.FindByFlagName(featureName, &actually)
		require.Nil(t, err)
		require.False(t, ok)
	})

	t.Run(`given we have a feature flag already set`, func(t *testing.T) {
		setup(t)

		ff := &rollouts.FeatureFlag{Name: featureName}
		require.Nil(t, spec.Subject.Save(ff))

		t.Run(`then searching for it returns the flag entity`, func(t *testing.T) {
			var actually rollouts.FeatureFlag

			found, err := spec.Subject.FindByFlagName(ff.Name, &actually)
			require.Nil(t, err)
			require.True(t, found)
			require.Equal(t, ff, &actually)
		})

	})
}
