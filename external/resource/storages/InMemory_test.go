package storages_test

import (
	"testing"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/external/resource/storages"
	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases"
	"github.com/toggler-io/toggler/usecases/specs"
)

var (
	_ usecases.Storage   = &storages.InMemory{}
	_ release.Storage    = &storages.InMemory{}
	_ security.Storage   = &storages.InMemory{}
	_ deployment.Storage = &storages.InMemory{}
)

func TestInMemory(t *testing.T) {
	specs.Storage{
		Subject:        storages.NewInMemory(),
		FixtureFactory: NewFixtureFactory(),
	}.Test(t)
}

func BenchmarkInMemory(b *testing.B) {
	specs.Storage{
		Subject:        storages.NewInMemory(),
		FixtureFactory: NewFixtureFactory(),
	}.Benchmark(b)
}

func TestNewTestingStorage(t *testing.T) {
	specs.Storage{
		Subject:        storages.NewTestingStorage(),
		FixtureFactory: NewFixtureFactory(),
	}.Test(t)
}
