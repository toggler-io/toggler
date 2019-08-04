package nullcache_test

import (
	"github.com/adamluzsi/toggler/extintf/caches"
	"github.com/adamluzsi/toggler/extintf/caches/nullcache"
	"github.com/adamluzsi/toggler/extintf/storages/inmemory"
	testing2 "github.com/adamluzsi/toggler/testing"
	"github.com/adamluzsi/toggler/usecases/specs"
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