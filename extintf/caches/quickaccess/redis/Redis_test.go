package redis_test

import (
	"os"
	"testing"

	"github.com/toggler-io/toggler/extintf/caches"
	"github.com/toggler-io/toggler/extintf/caches/cachespecs"
	"github.com/toggler-io/toggler/extintf/caches/quickaccess/redis"
	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases"
	"github.com/stretchr/testify/require"
)

func TestRedis(t *testing.T) {
	factory := func(s usecases.Storage) caches.Interface {
		cache, err := redis.New(getTestRedisConnstr(t), s)
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
