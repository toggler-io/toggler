package specs

import (
	testing2 "github.com/adamluzsi/FeatureFlags/testing"
	"testing"

	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/frameless/resources/specs"
)

type StorageSpec struct {
	Storage rollouts.Storage
}

func (spec *StorageSpec) Test(t *testing.T) {

	entityTypes := []interface{}{
		rollouts.FeatureFlag{},
		rollouts.Pilot{},
	}

	ff := testing2.NewFixtureFactory()

	for _, entityType := range entityTypes {
		specs.TestMinimumRequirements(t, spec.Storage, entityType, ff)
		specs.TestUpdate(t, spec.Storage, entityType, ff)
		specs.TestFindAll(t, spec.Storage, entityType, ff)
	}

	FlagFinderSpec{Subject: spec.Storage}.Test(t)
	PilotFinderSpec{Subject: spec.Storage}.Test(t)

}
