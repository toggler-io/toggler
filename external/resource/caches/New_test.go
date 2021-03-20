package caches_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/resource/caches"
	sh "github.com/toggler-io/toggler/spechelper"
)

func TestNew(t *testing.T) {
	s := testcase.NewSpec(t)
	sh.SetUp(s)

	s.Describe(`New`, func(s *testcase.Spec) {
		var (
			connstr = testcase.Var{Name: `connstr`}
			subject = func(t *testcase.T) (toggler.Storage, error) {
				cache, err := caches.New(connstr.Get(t).(string), sh.StorageGet(t))
				if err == nil {
					t.Defer(cache.Close)
				}
				return cache, err
			}
		)


		onSuccess := func(t *testcase.T) toggler.Storage {
			s, err := subject(t)
			require.Nil(t, err)
			return s
		}

		s.When(`the connection string is valid url but no implementation found`, func(s *testcase.Spec) {
			connstr.Let(s, func(t *testcase.T) interface{} { return `nexthypedstoragesystem://user:pwd@localhost:8100/db` })

			s.Then(`it will return null object implementation`, func(t *testcase.T) {
				require.Equal(t, sh.StorageGet(t), onSuccess(t))
			})
		})

		s.When(`the connection string is some Storage specific custom connstring`, func(s *testcase.Spec) {
			connstr.Let(s, func(t *testcase.T) interface{} { return `db=42 host=the-answer.com` })

			s.Then(`it will return null object implementation`, func(t *testcase.T) {
				require.Equal(t, sh.StorageGet(t), onSuccess(t))
			})
		})

		s.When(`the connstr is a symbolic name to use in "memory" caching`, func(s *testcase.Spec) {
			connstr.Let(s, func(t *testcase.T) interface{} { return `memory` })

			s.Then(`it will return the memory Storage object`, func(t *testcase.T) {
				s, isThat := onSuccess(t).(*caches.Manager)
				require.True(t, isThat)
				require.Nil(t, s.Close())
				_, ok := s.CacheStorage.(*caches.InMemoryCacheStorage)
				require.True(t, ok)
			})
		})

		s.When(`the connstr points to redis schema`, func(s *testcase.Spec) {
			connstr.Let(s, func(t *testcase.T) interface{} { return getTestRedisConnstr(t) })

			s.Then(`it will return the redis Storage object`, func(t *testcase.T) {
				s, isThat := onSuccess(t).(*caches.Manager)
				require.True(t, isThat)
				require.Nil(t, s.Close())
				_, ok := s.CacheStorage.(*caches.RedisCacheStorage)
				require.True(t, ok)
			})
		})
	})
}
