package specs

import (
	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/toggler/services/security"
	"testing"
)

type StorageSpec struct {
	Storage security.Storage
	resources.FixtureFactory
}

func (spec StorageSpec) Test(t *testing.T) {
	entityTypes := []interface{}{
		security.Token{},
	}

	for _, entityType := range entityTypes {
		resources.TestMinimumRequirements(t, spec.Storage, entityType, spec.FixtureFactory)
		resources.TestUpdate(t, spec.Storage, entityType, spec.FixtureFactory)
	}

	TokenFinderSpec{Subject: spec.Storage, FixtureFactory: spec.FixtureFactory}.Test(t)
}
