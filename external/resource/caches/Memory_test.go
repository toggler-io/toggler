package caches_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/external/resource/caches"
	"github.com/toggler-io/toggler/external/resource/caches/contracts"
	sh "github.com/toggler-io/toggler/spechelper"
)

func TestInMemoryCacheStorage(t *testing.T) {
	SpecInMemoryCacheStorage(t)
}

func BenchmarkInMemoryCacheStorage(b *testing.B) {
	SpecInMemoryCacheStorage(b)
}

func SpecInMemoryCacheStorage(tb testing.TB) {
	testcase.RunContract(tb, contracts.Storage{
		Subject: func(tb testing.TB) caches.Storage {
			return caches.NewInMemoryCacheStorage()
		},
		FixtureFactory: sh.DefaultFixtureFactory,
	})
}
