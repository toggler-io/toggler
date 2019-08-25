package inmemory_test

import (
	"testing"

	"github.com/toggler-io/toggler/extintf/storages/inmemory"
	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases/specs"
)

func TestInMemory(t *testing.T) {
	specs.StorageSpec{
		Subject:        inmemory.New(),
		FixtureFactory: NewFixtureFactory(),
	}.Test(t)
}

func BenchmarkInMemory(b *testing.B) {
	specs.StorageSpec{
		Subject:        inmemory.New(),
		FixtureFactory: NewFixtureFactory(),
	}.Benchmark(b)
}
