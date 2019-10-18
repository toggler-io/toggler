package caches_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/extintf/caches"
	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases"
)

func TestNew(t *testing.T) {
	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)

	s.Describe(`New`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (usecases.Storage, error) {
			return caches.New(t.I(`connstr`).(string), GetStorage(t))
		}

		onSuccess := func(t *testcase.T) usecases.Storage {
			s, err := subject(t)
			require.Nil(t, err)
			return s
		}

		s.When(`the connection string is valid url but no implementation found`, func(s *testcase.Spec) {
			s.Let(`connstr`, func(t *testcase.T) interface{} { return `nexthypedstoragesystem://user:pwd@localhost:8100/db` })

			s.Then(`then it will return null object implementation`, func(t *testcase.T) {
				_, isThat := onSuccess(t).(*caches.NullCache)

				require.True(t, isThat)
			})
		})

		s.When(`the connection string is some Storage specific custom connstring`, func(s *testcase.Spec) {
			s.Let(`connstr`, func(t *testcase.T) interface{} { return `db=42 host=the-answer.com` })

			s.Then(`then it will return null object implementation`, func(t *testcase.T) {
				_, isThat := onSuccess(t).(*caches.NullCache)

				require.True(t, isThat)
			})
		})
	})
}
