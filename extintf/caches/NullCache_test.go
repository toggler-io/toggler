package caches_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/extintf/caches"
	"github.com/toggler-io/toggler/extintf/storages"
	testing2 "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases/specs"
)

func TestNullCache(t *testing.T) {
	s := storages.NewInMemory()
	c := caches.NewNullCache(s)
	specs.StorageSpec{Subject: c, FixtureFactory: testing2.NewFixtureFactory()}.Test(t)
}

func TestNullCacheImpCacheInterface(t *testing.T) {
	s := storages.NewInMemory()
	c := caches.NewNullCache(s)
	var _ caches.Interface = c
	require.Nil(t, c.SetTimeToLiveForValuesToCache(42))
}