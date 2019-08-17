package specs

import (
	"github.com/adamluzsi/frameless/resources"
	rollspecs "github.com/adamluzsi/toggler/services/rollouts/specs"
	secuspecs "github.com/adamluzsi/toggler/services/security/specs"
	"testing"

	"github.com/adamluzsi/toggler/usecases"
)

type StorageSpec struct {
	Subject usecases.Storage

	FixtureFactory interface {
		resources.FixtureFactory
		SetPilotFeatureFlagID(ffID string) func()
	}
}

func (spec StorageSpec) Test(t *testing.T) {
	rollspecs.StorageSpec{Storage: spec.Subject, FixtureFactory: spec.FixtureFactory}.Test(t)
	secuspecs.StorageSpec{Storage: spec.Subject, FixtureFactory: spec.FixtureFactory}.Test(t)
}
