package storages_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/external/resource/storages"
	testing2 "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases/specs"
)

func BenchmarkRedis(b *testing.B) {
	r, err := storages.NewRedis(getTestRedisConnstr(b))
	require.Nil(b, err)
	specs.StorageSpec{
		Subject:        r,
		FixtureFactory: testing2.NewFixtureFactory(),
	}.Benchmark(b)
}

func TestRedis(t *testing.T) {
	r, err := storages.NewRedis(getTestRedisConnstr(t))
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
