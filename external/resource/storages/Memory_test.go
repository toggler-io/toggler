package storages_test

import (
	"testing"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/domains/toggler/contracts"
	"github.com/toggler-io/toggler/external/resource/storages"
	sh "github.com/toggler-io/toggler/spechelper"
)

var (
	_ toggler.Storage    = &storages.Memory{}
	_ release.Storage    = &storages.Memory{}
	_ security.Storage   = &storages.Memory{}
	_ deployment.Storage = &storages.Memory{}
)

func TestMemory(t *testing.T) {
	contracts.Storage{
		Subject: func(tb testing.TB) toggler.Storage {
			return storages.NewMemory()
		},
		FixtureFactory: sh.DefaultFixtureFactory,
	}.Test(t)
}

func BenchmarkMemory(b *testing.B) {
	contracts.Storage{
		Subject: func(tb testing.TB) toggler.Storage {
			return storages.NewMemory()
		},
		FixtureFactory: sh.DefaultFixtureFactory,
	}.Benchmark(b)
}
