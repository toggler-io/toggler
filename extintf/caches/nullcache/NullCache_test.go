package nullcache_test

import (
	"github.com/adamluzsi/toggler/extintf/caches/nullcache"
	"github.com/adamluzsi/toggler/extintf/storages/inmemory"
	"github.com/adamluzsi/toggler/usecases/specs"
	"testing"
)

func TestNullCache(t *testing.T) {
	s := inmemory.New()
	c := nullcache.NewNullCache(s)
	specs.StorageSpec{Subject: c}.Test(t)
}