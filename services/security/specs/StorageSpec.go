package specs

import (
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/toggler/services/security"
	testing2 "github.com/adamluzsi/toggler/testing"
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

	ff := testing2.NewFixtureFactory()

	for _, entityType := range entityTypes {
		specs.TestMinimumRequirements(t, spec.Storage, entityType, ff)
		specs.TestUpdate(t, spec.Storage, entityType, ff)
	}

	TokenFinderSpec{Subject: spec.Storage, FixtureFactory: spec.FixtureFactory}.Test(t)
}
