package testing_test

import (
	tting "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases/specs"
	"testing"
)

func TestTestStorage(t *testing.T) {
	specs.StorageSpec{
		Subject: tting.NewTestStorage(),
		FixtureFactory: tting.NewFixtureFactory(),
	}.Test(t)
}
