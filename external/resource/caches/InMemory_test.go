package caches_test

import (
	"testing"

	"github.com/toggler-io/toggler/external/resource/caches"
	"github.com/toggler-io/toggler/external/resource/caches/cachespecs"
	"github.com/toggler-io/toggler/external/resource/storages"
	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases"
	"github.com/toggler-io/toggler/usecases/specs"
)

func TestInMemory_StorageSpec(t *testing.T) {
	specs.StorageSpec{
		Subject:        caches.NewInMemory(storages.NewInMemory()),
		FixtureFactory: NewFixtureFactory(),
	}.Test(t)
}

func TestInMemory_CacheSpec(t *testing.T) {
	cachespecs.CacheSpec{
		Factory:        func(s usecases.Storage) caches.Interface {
			return caches.NewInMemory(s)
		},
		FixtureFactory: NewFixtureFactory(),
	}.Test(t)
}

func BenchmarkInMemory_CacheSpec(b *testing.B) {
	cachespecs.CacheSpec{
		Factory:        func(s usecases.Storage) caches.Interface {
			return caches.NewInMemory(s)
		},
		FixtureFactory: NewFixtureFactory(),
	}.Benchmark(b)
}