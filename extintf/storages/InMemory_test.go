package storages_test

import (
	"testing"

	"github.com/toggler-io/toggler/extintf/storages"
	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases/specs"
)

func TestInMemory(t *testing.T) {
	specs.StorageSpec{
		Subject:        storages.NewInMemory(),
		FixtureFactory: NewFixtureFactory(),
	}.Test(t)
}

func BenchmarkInMemory(b *testing.B) {
	specs.StorageSpec{
		Subject:        storages.NewInMemory(),
		FixtureFactory: NewFixtureFactory(),
	}.Benchmark(b)
}
