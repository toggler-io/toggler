package storages_test

import (
	"fmt"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/toggler/extintf/storages"
	"github.com/adamluzsi/toggler/extintf/storages/inmemory"
	"github.com/adamluzsi/toggler/extintf/storages/postgres"
	"github.com/adamluzsi/toggler/usecases"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`New`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (usecases.Storage, error) {
			return storages.New(t.I(`connstr`).(string))
		}

		onSuccess := func(t *testcase.T) usecases.Storage {
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

		s.When(`the connection string is a postgres db`, func(s *testcase.Spec) {
			s.Let(`connstr`, func(t *testcase.T) interface{} {
				return fmt.Sprintf(`%s://postgres@localhost:8100/postgres?sslmode=disable`, t.I(`driver name`).(string))
			})

			for _, driverName := range []string{
				`postgres`,
			} {
				s.And(fmt.Sprintf(`with the driver name is %s`, driverName), func(s *testcase.Spec) {
					driverName := driverName // pass by value copy to remove loop variable dependency in the callbacks
					s.Let(`driver name`, func(t *testcase.T) interface{} { return driverName })

					s.Then(`then it will return postgres implementation`, func(t *testcase.T) {
						_, isPG := onSuccess(t).(*postgres.Postgres)

						require.True(t, isPG)
					})

				})
			}

		})

		s.When(`the connection string is a "memory"`, func(s *testcase.Spec) {
			s.Let(`connstr`, func(t *testcase.T) interface{} { return `memory` })

			s.Then(`then it will return "inmemory" implementation`, func(t *testcase.T) {
				_, isThat := onSuccess(t).(*inmemory.InMemory)

				require.True(t, isThat)
			})
		})
	})
}
