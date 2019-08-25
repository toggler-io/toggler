package nullcache_test

import (
	"github.com/toggler-io/toggler/extintf/caches"
	"github.com/toggler-io/toggler/extintf/caches/nullcache"
	"github.com/toggler-io/toggler/extintf/storages/inmemory"
	testing2 "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases/specs"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNullCache(t *testing.T) {
	s := inmemory.New()
	c := nullcache.NewNullCache(s)
	specs.StorageSpec{Subject: c, FixtureFactory: testing2.NewFixtureFactory()}.Test(t)
}

func TestNullCacheImpCacheInterface(t *testing.T) {
	s := inmemory.New()
	c := nullcache.NewNullCache(s)
	var _ caches.Interface = c
	require.Nil(t, c.SetTimeToLiveForValuesToCache(42))
}