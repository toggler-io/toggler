package testing_test

import (
	tting "github.com/adamluzsi/toggler/testing"
	"github.com/adamluzsi/toggler/usecases/specs"
	"testing"
)

func TestTestStorage(t *testing.T) {
	(&specs.StorageSpec{Storage:tting.NewTestStorage()}).Test(t)
}
