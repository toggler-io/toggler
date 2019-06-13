package specs

import (
	rollspecs "github.com/adamluzsi/toggler/services/rollouts/specs"
	secuspecs "github.com/adamluzsi/toggler/services/security/specs"

	"testing"

	"github.com/adamluzsi/toggler/usecases"
)

type StorageSpec struct {
	Storage usecases.Storage
}

func (spec *StorageSpec) Test(t *testing.T) {
	(&rollspecs.StorageSpec{Storage: spec.Storage}).Test(t)
	(&secuspecs.StorageSpec{Storage: spec.Storage}).Test(t)
}
