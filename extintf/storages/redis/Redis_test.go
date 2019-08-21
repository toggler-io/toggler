package redis_test

import (
	"github.com/adamluzsi/toggler/extintf/storages/redis"
	testing2 "github.com/adamluzsi/toggler/testing"
	"github.com/adamluzsi/toggler/usecases/specs"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func BenchmarkRedis(b *testing.B) {
	r, err := redis.New(getTestRedisConnstr(b))
	require.Nil(b, err)
	specs.StorageSpec{
		Subject:        r,
		FixtureFactory: testing2.NewFixtureFactory(),
	}.Benchmark(b)
}

func TestRedis(t *testing.T) {
	r, err := redis.New(getTestRedisConnstr(t))
	require.Nil(t, err)
	specs.StorageSpec{
		Subject:        r,
		FixtureFactory: testing2.NewFixtureFactory(),
	}.Test(t)
}

func getTestRedisConnstr(tb testing.TB) string {
	value, isSet := os.LookupEnv(`TEST_STORAGE_URL_REDIS`)

	if !isSet {
		tb.Skip(`redis url is not set in "TEST_STORAGE_URL_REDIS"`)
	}

	return value
}
