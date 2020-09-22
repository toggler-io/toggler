package caches_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/toggler/specs"
	"github.com/toggler-io/toggler/external/resource/caches"
	"github.com/toggler-io/toggler/external/resource/storages"
	testing2 "github.com/toggler-io/toggler/testing"
)

func TestNullCache(t *testing.T) {
	s := storages.NewInMemory()
	c := caches.NewNullCache(s)
	specs.Storage{Subject: c, FixtureFactory: testing2.NewFixtureFactory()}.Test(t)
}

func TestNullCacheImpCacheInterface(t *testing.T) {
	s := storages.NewInMemory()
	c := caches.NewNullCache(s)
	var _ caches.Interface = c
	require.Nil(t, c.SetTimeToLiveForValuesToCache(42))
}