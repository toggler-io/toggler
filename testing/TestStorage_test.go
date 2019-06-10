package testing_test

import (
	rollspecs "github.com/adamluzsi/toggler/services/rollouts/specs"
	secuspecs "github.com/adamluzsi/toggler/services/security/specs"
	tting "github.com/adamluzsi/toggler/testing"
	"testing"
)

func TestTestStorage(t *testing.T) {
	//t.Skip(`TODO finish`)
	storage := tting.NewTestStorage()
	(&rollspecs.StorageSpec{Storage: storage}).Test(t)
	(&secuspecs.StorageSpec{Storage: storage}).Test(t)
}
