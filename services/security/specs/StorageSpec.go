package specs

import (
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/toggler/services/security"
	"testing"
)

type StorageSpec struct {
	Storage security.Storage
	specs.FixtureFactory
}

func (spec StorageSpec) Test(t *testing.T) {
	entityTypes := []interface{}{
		security.Token{},
	}

	for _, entityType := range entityTypes {
		specs.TestMinimumRequirements(t, spec.Storage, entityType, spec.FixtureFactory)
		specs.TestUpdate(t, spec.Storage, entityType, spec.FixtureFactory)
	}

	TokenFinderSpec{Subject: spec.Storage, FixtureFactory: spec.FixtureFactory}.Test(t)
}
