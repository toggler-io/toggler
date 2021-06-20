package caches_test

//import (
//	"os"
//	"testing"
//
//	"github.com/adamluzsi/testcase"
//	"github.com/stretchr/testify/require"
//	"github.com/toggler-io/toggler/external/resource/caches"
//	"github.com/toggler-io/toggler/external/resource/caches/contracts"
//	sh "github.com/toggler-io/toggler/spechelper"
//)
//
//func TestRedisCacheStorage(t *testing.T) {
//	SpecRedisCacheStorage(t)
//}
//
//func BenchmarkRedisCacheStorage(b *testing.B) {
//	SpecRedisCacheStorage(b)
//}
//
//func SpecRedisCacheStorage(tb testing.TB) {
//	connstr := getTestRedisConnstr(tb)
//	testcase.RunContract(tb, contracts.Storage{
//		Subject: func(tb testing.TB) caches.storage {
//			cs, err := caches.NewRedisCacheStorage(connstr)
//			require.Nil(tb, err)
//			return cs
//		},
//		FixtureFactory: sh.DefaultFixtureFactory,
//	})
//}
//
//func getTestRedisConnstr(tb testing.TB) string {
//	const envKey = `TEST_CACHE_URL_REDIS`
//	value, isSet := os.LookupEnv(envKey)
//	if !isSet {
//		tb.Skipf(`redis url is not index in "%s"`, envKey)
//	}
//
//	return value
//}
