package storages_test

import (
	"os"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/resource/storages"
)

func TestNew(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`New`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (toggler.Storage, error) {
			return storages.New(t.I(`connstr`).(string))
		}

		onSuccess := func(t *testcase.T) toggler.Storage {
			s, err := subject(t)
			require.Nil(t, err)
			return s
		}

		s.When(`the connection string is unknown`, func(s *testcase.Spec) {
			s.Let(`connstr`, func(t *testcase.T) interface{} {
				return `nexthypedstoragesystem://user:pwd@localhost:8100/db`
			})

			s.Then(`it will result in error`, func(t *testcase.T) {
				s, err := subject(t)
				require.Nil(t, s)
				require.Error(t, err)
			})
		})

		s.When(`the connection string is a "postgres"`, func(s *testcase.Spec) {
			s.Let(`connstr`, func(t *testcase.T) interface{} {
				return os.Getenv(`TEST_DATABASE_URL_POSTGRES`)
			})

			s.Then(`then it will return postgres implementation`, func(t *testcase.T) {
				_, isPG := onSuccess(t).(*storages.Postgres)

				require.True(t, isPG)
			})
		})

		s.When(`the connection string is a "memory"`, func(s *testcase.Spec) {
			s.Let(`connstr`, func(t *testcase.T) interface{} { return `memory` })

			s.Then(`then it will return "inmemory" implementation`, func(t *testcase.T) {
				_, isThat := onSuccess(t).(*storages.InMemory)

				require.True(t, isThat)
			})
		})
	})
}
