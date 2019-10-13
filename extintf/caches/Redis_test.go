package caches_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/extintf/caches"
	"github.com/toggler-io/toggler/extintf/caches/cachespecs"
	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases"
)

func TestRedis(t *testing.T) {
	factory := func(s usecases.Storage) caches.Interface {
		cache, err := caches.NewRedis(getTestRedisConnstr(t), s)
		require.Nil(t, err)
		return cache
	}

	cachespecs.CacheSpec{
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
