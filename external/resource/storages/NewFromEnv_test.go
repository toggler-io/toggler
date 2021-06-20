package storages_test

import (
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/resource/storages"
)

func TestNewFromEnv(t *testing.T) {
	s := testcase.NewSpec(t)

	subject := func() (toggler.Storage, error) {
		return storages.NewFromEnv()
	}

	s.Around(func(t *testcase.T) func() {
		defers := []func(){
			UnsetEnv(t, `DATABASE_URL`),
			UnsetEnv(t, `RDS_PORT`),
			UnsetEnv(t, `RDS_HOSTNAME`),
			UnsetEnv(t, `RDS_USERNAME`),
			UnsetEnv(t, `RDS_PASSWORD`),
			UnsetEnv(t, `RDS_DB_NAME`),
		}

		return func() {
			for _, d := range defers {
				d()
			}
		}
	})

	s.Then(`it will return with error that environment lack values for storage initialization`, func(t *testcase.T) {
		_, err := subject()
		require.Equal(t, storages.ErrNewFromErrNotPossible, err)
	})

	s.When(`DATABASE_URL is set`, func(s *testcase.Spec) {
		s.Around(func(t *testcase.T) func() {
			return SetEnv(t, `DATABASE_URL`, `memory`)
		})

		s.Then(`it will return the storage initialized by the DATABASE_URL value`, func(t *testcase.T) {
			storage, err := subject()
			require.Nil(t, err)
			require.NotNil(t, storage)
			_, ok := storage.(*storages.Memory)
			require.True(t, ok)
		})
	})

	s.When(`Amazon Relational Database Service variables are set`, func(s *testcase.Spec) {
		s.Let(`pg_url`, func(t *testcase.T) interface{} {
			u, err := url.Parse(os.Getenv(`TEST_DATABASE_URL_POSTGRES`))
			require.Nil(t, err)
			return u
		})

		s.Around(func(t *testcase.T) func() {
			u := t.I(`pg_url`).(*url.URL)
			_, dbName := path.Split(u.Path)
			passwd, _ := u.User.Password()

			defers := []func(){
				SetEnv(t, `RDS_PORT`, u.Port()),
				SetEnv(t, `RDS_HOSTNAME`, u.Hostname()),
				SetEnv(t, `RDS_USERNAME`, u.User.Username()),
				SetEnv(t, `RDS_PASSWORD`, passwd),
				SetEnv(t, `RDS_DB_NAME`, dbName),
			}

			return func() {
				for _, d := range defers {
					d()
				}
			}
		})

		s.And(`RDS_ENGINE is set`, func(s *testcase.Spec) {
			s.Around(func(t *testcase.T) func() {
				return SetEnv(t, `RDS_ENGINE`, `postgres`)
			})

			shouldBePostgresStorage := func(t *testcase.T, s toggler.Storage) {
				require.NotNil(t, s)
				_, ok := s.(*storages.Postgres)
				require.True(t, ok)
			}

			s.And(`RDS_ENGINE_OPTS also set`, func(s *testcase.Spec) {
				s.Around(func(t *testcase.T) func() {
					return SetEnv(t, `RDS_ENGINE_OPTS`, t.I(`pg_url`).(*url.URL).RawQuery)
				})

				s.Then(`it will build the proper database url`, func(t *testcase.T) {
					s, err := subject()
					require.Nil(t, err)
					shouldBePostgresStorage(t, s)
				})
			})

			s.And(`RDS_ENGINE_OPTS is unset`, func(s *testcase.Spec) {
				s.Around(func(t *testcase.T) func() {
					return UnsetEnv(t, `RDS_ENGINE_OPTS`)
				})

				s.Then(`it will build the proper database url without the extra settings`, func(t *testcase.T) {
					s, err := subject()
					if err != nil {
						require.Equal(t, pq.ErrSSLNotSupported, err)
						return
					}
					shouldBePostgresStorage(t, s)
				})
			})
		})

		s.And(`RDS_ENGINE is undefined`, func(s *testcase.Spec) {
			s.Around(func(t *testcase.T) func() {
				return UnsetEnv(t, `RDS_ENGINE`)
			})

			s.Then(`it will return with error that explain this`, func(t *testcase.T) {
				_, err := subject()
				require.Equal(t, storages.ErrRDSEngineNotSet, err)
			})
		})
	})
}

func UnsetEnv(t testing.TB, key string) func() {
	current, isSet := os.LookupEnv(key)
	if !isSet {
		return func() {}
	}

	require.Nil(t, os.Unsetenv(key))
	return func() { require.Nil(t, os.Setenv(key, current)) }
}

func SetEnv(t testing.TB, key, value string) func() {
	current, isSet := os.LookupEnv(key)
	require.Nil(t, os.Setenv(key, value))
	return func() {
		if isSet {
			require.Nil(t, os.Setenv(key, current))
		} else {
			require.Nil(t, os.Unsetenv(key))
		}
	}
}
