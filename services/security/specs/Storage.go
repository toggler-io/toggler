package specs

import (
	"github.com/adamluzsi/FeatureFlags/services/security"
	testing2 "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/frameless/resources/specs"
	"testing"
)

type StorageSpec struct {
	Storage security.Storage
}

func (spec *StorageSpec) Test(t *testing.T) {

	entityTypes := []interface{}{
		security.Token{},
	}

	ff := testing2.NewFixtureFactory()

	for _, entityType := range entityTypes {
		specs.TestMinimumRequirements(t, spec.Storage, entityType, ff)
		specs.TestUpdate(t, spec.Storage, entityType, ff)
	}

	TokenFinderSpec{Subject: spec.Storage}.Test(t)

}
