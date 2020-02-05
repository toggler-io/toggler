package storages_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/extintf/storages"
	testing2 "github.com/toggler-io/toggler/testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/frameless/resources"

	"github.com/toggler-io/toggler/usecases/specs"
)

func BenchmarkPostgres(b *testing.B) {
	if testing.Short() {
		b.Skip()
	}

	db := MustOpenDB(b)
	defer db.Close()

	storage, err := storages.NewPostgres(db)
	require.Nil(b, err)

	specs.StorageSpec{
		Subject:        storage,
		FixtureFactory: testing2.NewFixtureFactory(),
	}.Benchmark(b)
}

func TestPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := MustOpenDB(t)
	defer db.Close()

	storage, err := storages.NewPostgres(db)
	require.Nil(t, err)

	specs.StorageSpec{
		Subject:        storage,
		FixtureFactory: testing2.NewFixtureFactory(),
	}.Test(t)
}

func MustOpenDB(tb testing.TB) *sql.DB {
	databaseConnectionString := getDatabaseConnectionString(tb)
	db, err := sql.Open("postgres", databaseConnectionString)
	require.Nil(tb, err)
	require.Nil(tb, db.Ping())
	return db
}

func getDatabaseConnectionString(tb testing.TB) string {
	databaseURL, isSet := os.LookupEnv("TEST_STORAGE_URL_POSTGRES")

	if !isSet {
		tb.Skip(`"TEST_STORAGE_URL_POSTGRES" env var is not set, therefore skipping this test`)
	}

	return databaseURL
}

func TestPostgres_Close(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	s := testcase.NewSpec(t)

	pg := func(t *testcase.T) *storages.Postgres {
		return &storages.Postgres{DB: t.I(`DB`).(interface{
			ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
			QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
			QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
		})}
	}

	subject := func(t *testcase.T) error {
		return pg(t).Close()
	}

	s.Let(`*sql.DB`, func(t *testcase.T) interface{} { return MustOpenDB(t.T) })
	s.After(func(t *testcase.T) { _ = t.I(`*sql.DB`).(*sql.DB).Close() })

	s.When(`db is a *sql.DB`, func(s *testcase.Spec) {
		s.Let(`DB`, func(t *testcase.T) interface{} { return t.I(`*sql.DB`) })

		s.Then(`it will close the *sql.DB object`, func(t *testcase.T) {
			require.Nil(t, subject(t))

			sqlDB := t.I(`DB`).(*sql.DB)
			row := sqlDB.QueryRow(`SELECT 1=1`)
			var v sql.NullBool
			err := row.Scan(&v)
			require.Error(t, err)
			require.Contains(t, err.Error(), `closed`)
		})
	})

	s.When(`db is a *sql.Tx`, func(s *testcase.Spec) {
		s.Let(`DB`, func(t *testcase.T) interface{} {
			tx, err := t.I(`*sql.DB`).(*sql.DB).Begin()
			require.Nil(t, err)
			return tx
		})

		s.Then(`the *sql.Tx will be in a Done state`, func(t *testcase.T) {
			require.Nil(t, subject(t))

			tx := t.I(`DB`).(*sql.Tx)
			row := tx.QueryRow(`SELECT 1=1`)
			var v sql.NullBool
			err := row.Scan(&v)
			require.Error(t, err)
			require.Equal(t, sql.ErrTxDone, err)
		})

		s.Then(`the *sql.Tx had received Commit`, func(t *testcase.T) {
			var te resources.TestEntity
			ctx := context.Background()

			pgSqlDB := &storages.Postgres{DB: t.I(`*sql.DB`).(*sql.DB)}
			require.Nil(t, pgSqlDB.Truncate(ctx, te))
			pgSqlTx := pg(t)

			require.Nil(t, pgSqlTx.Create(ctx, &te))
			require.Nil(t, pgSqlTx.Close())

			count, err := iterators.Count(pgSqlDB.FindAll(ctx, te))
			require.Nil(t, err)
			require.Equal(t, 1, count)
		})
	})

}
