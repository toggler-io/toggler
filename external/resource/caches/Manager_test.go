package caches_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/resource/caches"
	"github.com/toggler-io/toggler/external/resource/caches/contracts"
	sh "github.com/toggler-io/toggler/spechelper"
)

func TestManager(t *testing.T) {
	SpecManager(t)
}

func BenchmarkMemory(b *testing.B) {
	SpecManager(b)
}

func SpecManager(tb testing.TB) {
	s := testcase.NewSpec(tb)
	sh.SetUp(s)

	s.Test(`contracts.Storage`, func(t *testcase.T) {
		testcase.RunContract(t, contracts.Storage{
			Subject:        NewFakeCacheStorage,
			FixtureFactory: sh.DefaultFixtureFactory,
		})
	})

	s.Test(`contracts.Cache`, func(t *testcase.T) {
		testcase.RunContract(t, contracts.Cache{
			NewCache: func(tb testing.TB, s toggler.Storage) toggler.Storage {
				m, err := caches.NewManager(s, NewFakeCacheStorage(tb))
				require.Nil(tb, err)
				return m
			},
		})
	})
}

func NewFakeCacheStorage(testing.TB) caches.Storage {
	imcs := caches.NewInMemoryCacheStorage()
	// sync subscription handling should ensure that logic that based on callback deterministically reliable.
	imcs.Options.DisableAsyncSubscriptionHandling = true
	return imcs
}
