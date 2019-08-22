package redis_test

import (
	"os"
	"testing"

	"github.com/adamluzsi/toggler/extintf/caches"
	"github.com/adamluzsi/toggler/extintf/caches/cachespecs"
	"github.com/adamluzsi/toggler/extintf/caches/quickaccess/redis"
	. "github.com/adamluzsi/toggler/testing"
	"github.com/adamluzsi/toggler/usecases"
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
