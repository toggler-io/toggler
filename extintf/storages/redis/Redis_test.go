package redis_test

import (
	"github.com/adamluzsi/toggler/extintf/storages/redis"
	testing2 "github.com/adamluzsi/toggler/testing"
	"github.com/adamluzsi/toggler/usecases/specs"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestRedis(t *testing.T) {
	r, err := redis.New(getTestRedisConnstr(t))
	require.Nil(t, err)
	specs.StorageSpec{
		Subject:        r,
		FixtureFactory: testing2.NewFixtureFactory(),
	}.Test(t)
}

func getTestRedisConnstr(t *testing.T) string {
	value, isSet := os.LookupEnv(`TEST_STORAGE_URL_REDIS`)

	if !isSet {
		t.Skip(`redis url is not set in "TEST_STORAGE_URL_REDIS"`)
	}

	return value
}
