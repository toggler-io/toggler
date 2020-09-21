package storages_test

import (
	"testing"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/external/resource/storages"
	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases/specs"
)

var (
	_ release.Storage    = &storages.InMemory{}
	_ security.Storage   = &storages.InMemory{}
	_ deployment.Storage = &storages.InMemory{}
)

func TestDevStorage(t *testing.T) {
	specs.Storage{
		Subject:        storages.NewDevStorage(),
		FixtureFactory: NewFixtureFactory(),
	}.Test(t)
}

func BenchmarkDevStorage(b *testing.B) {
	specs.Storage{
		Subject:        storages.NewDevStorage(),
		FixtureFactory: NewFixtureFactory(),
	}.Benchmark(b)
}

