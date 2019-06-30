package specs

import (
	rollspecs "github.com/adamluzsi/toggler/services/rollouts/specs"
	secuspecs "github.com/adamluzsi/toggler/services/security/specs"

	"testing"

	"github.com/adamluzsi/toggler/usecases"
)

type StorageSpec struct {
	Subject usecases.Storage
}

func (spec StorageSpec) Test(t *testing.T) {
	(rollspecs.StorageSpec{Storage: spec.Subject}).Test(t)
	(secuspecs.StorageSpec{Storage: spec.Subject}).Test(t)
}
