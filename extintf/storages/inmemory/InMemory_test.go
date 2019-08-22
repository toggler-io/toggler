package inmemory_test

import (
	"testing"

	"github.com/adamluzsi/toggler/extintf/storages/inmemory"
	. "github.com/adamluzsi/toggler/testing"
	"github.com/adamluzsi/toggler/usecases/specs"
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
