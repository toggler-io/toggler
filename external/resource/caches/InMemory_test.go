package caches_test

import (
	"testing"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/domains/toggler/specs"
	"github.com/toggler-io/toggler/external/resource/caches"
	cachespecs "github.com/toggler-io/toggler/external/resource/caches/specs"
	"github.com/toggler-io/toggler/external/resource/storages"
	. "github.com/toggler-io/toggler/testing"
)

func TestInMemory_StorageSpec(t *testing.T) {
	specs.Storage{
		Subject:        caches.NewInMemory(storages.NewInMemory()),
		FixtureFactory: DefaultFixtureFactory,
	}.Test(t)
}

func TestInMemory_CacheSpec(t *testing.T) {
	cachespecs.Cache{
		Factory: func(s toggler.Storage) caches.Interface {
			return caches.NewInMemory(s)
		},
		FixtureFactory: DefaultFixtureFactory,
	}.Test(t)
}

func BenchmarkInMemory_CacheSpec(b *testing.B) {
	cachespecs.Cache{
		Factory: func(s toggler.Storage) caches.Interface {
			return caches.NewInMemory(s)
		},
		FixtureFactory: DefaultFixtureFactory,
	}.Benchmark(b)
}
