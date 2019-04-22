package specs

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/frameless/resources/specs"
	"testing"
)

type StorageSpec struct {
	Storage rollouts.Storage
}

func (spec *StorageSpec) Test(t *testing.T) {

	entities := []interface{}{
		rollouts.FeatureFlag{},
		rollouts.Pilot{},
	}

	for _, entity := range entities {
		specs.TestMinimumRequirements(t, spec.Storage, entity)
	}

	FlagFinderSpec{Subject: spec.Storage}.Test(t)
	PilotFinderSpec{Subject: spec.Storage}.Test(t)

}