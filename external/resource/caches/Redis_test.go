package caches_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/resource/caches"
	"github.com/toggler-io/toggler/external/resource/caches/specs"
	. "github.com/toggler-io/toggler/testing"
)

func TestRedis(t *testing.T) {
	cache, err := caches.NewRedis(getTestRedisConnstr(t), nil)
	require.Nil(t, err)
	defer cache.Close()

	factory := func(s toggler.Storage) caches.Interface {
		cache.Storage = s
		return cache
	}

	specs.Cache{
		Factory:        factory,
		FixtureFactory: NewFixtureFactory(),
	}.Test(t)
}

func getTestRedisConnstr(t *testing.T) string {
	value, isSet := os.LookupEnv(`TEST_CACHE_URL_REDIS`)

	if !isSet {
		t.Skip(`redis url is not set in "TEST_CACHE_URL_REDIS"`)
	}

	return value
}
