package usecases_test

import (
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testrun"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestUseCases_IsFeatureGloballyEnabled(t *testing.T) {
	t.Parallel()

	var (
		featureFlagName string
		uc              *usecases.UseCases
		storage         usecases.Storage
	)

	subject := func() (bool, error) {
		return uc.IsFeatureGloballyEnabled(featureFlagName)
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
	})

	t.Run(`when flag is fully rolled out`, func(t *testing.T) {
		steps := steps.Add(func(t *testing.T) {
			require.Nil(t, uc.UpdateFeatureFlagRolloutPercentage(featureFlagName, 100))
		})

		t.Run(`then feature is enabled`, func(t *testing.T) {
			steps.Setup(t)

			require.True(t, isEnrolled(t))
		})
	})

	t.Run(`when flag is not yet fully rolled out`, func(t *testing.T) {
		steps := steps.Add(func(t *testing.T) {
			require.Nil(t, uc.UpdateFeatureFlagRolloutPercentage(featureFlagName, rand.Intn(100)))
		})

		t.Run(`then feature is enabled`, func(t *testing.T) {
			steps.Setup(t)

			require.False(t, isEnrolled(t))
		})
	})
}
