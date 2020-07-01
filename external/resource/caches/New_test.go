package caches_test

import (
	"os"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/external/resource/caches"
	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases"
)

func TestNew(t *testing.T) {
	s := testcase.NewSpec(t)
	SetUp(s)

	s.Describe(`New`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (usecases.Storage, error) {
			cache, err := caches.New(t.I(`connstr`).(string), ExampleStorage(t))
			if err == nil {
				t.Defer(cache.Close)
			}
			return cache, err
		}

		onSuccess := func(t *testcase.T) usecases.Storage {
			s, err := subject(t)
			require.Nil(t, err)
			return s
		}

		s.When(`the connection string is valid url but no implementation found`, func(s *testcase.Spec) {
			s.Let(`connstr`, func(t *testcase.T) interface{} { return `nexthypedstoragesystem://user:pwd@localhost:8100/db` })

			s.Then(`it will return null object implementation`, func(t *testcase.T) {
				_, isThat := onSuccess(t).(*caches.NullCache)

				require.True(t, isThat)
			})
		})

		s.When(`the connection string is some Storage specific custom connstring`, func(s *testcase.Spec) {
			s.Let(`connstr`, func(t *testcase.T) interface{} { return `db=42 host=the-answer.com` })

			s.Then(`it will return null object implementation`, func(t *testcase.T) {
				_, isThat := onSuccess(t).(*caches.NullCache)

				require.True(t, isThat)
			})
		})

		s.When(`the connection string belongs to a redis server`, func(s *testcase.Spec) {
			s.Let(`connstr`, func(t *testcase.T) interface{} { return os.Getenv(`TEST_CACHE_URL_REDIS`) })

			s.Then(`it will return the redis cache object`, func(t *testcase.T) {
				s, isThat := onSuccess(t).(*caches.Redis)
				require.True(t, isThat)
				require.Nil(t, s.Close())
			})
		})

		s.When(`the connstr is a symbolic name to use in "memory" caching`, func(s *testcase.Spec) {
			s.Let(`connstr`, func(t *testcase.T) interface{} { return `memory` })

			s.Then(`it will return the redis cache object`, func(t *testcase.T) {
				s, isThat := onSuccess(t).(*caches.InMemory)
				require.True(t, isThat)
				require.Nil(t, s.Close())
			})
		})

	})
}
