package storages_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/domains/toggler"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/external/resource/storages"

	sh "github.com/toggler-io/toggler/spechelper"

	"github.com/toggler-io/toggler/domains/toggler/contracts"
)

func BenchmarkPostgres(b *testing.B) {
	if testing.Short() {
		b.Skip()
	}

	db, dsn := MustOpenDB(b)
	defer db.Close()

	storage, err := storages.NewPostgres(db, dsn)
	require.Nil(b, err)

	contracts.Storage{
		Subject: func(tb testing.TB) toggler.Storage {
			return storage
		},
		FixtureFactory: sh.DefaultFixtureFactory,
	}.Benchmark(b)
}

func TestPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db, dsn := MustOpenDB(t)
	defer db.Close()

	storage, err := storages.NewPostgres(db, dsn)
	require.Nil(t, err)

	contracts.Storage{
		Subject: func(tb testing.TB) toggler.Storage {
			return storage
		},
		FixtureFactory: sh.DefaultFixtureFactory,
	}.Test(t)
}

func MustOpenDB(tb testing.TB) (*sql.DB, string) {
	// I don't know exactly how but somehow `DELETE` queries from different connections made in the past
	// might affect the results in this connection,
	// resulting that some of the data goes missing during tests.
	// To reproduce this, please execute full project testing suite with E2E mode, while removing this sleep.
	//
	// TODO: TECH-DEBT
	time.Sleep(time.Second)
	databaseConnectionString := getDatabaseConnectionString(tb)
	db, err := sql.Open("postgres", databaseConnectionString)
	require.Nil(tb, err)
	require.Nil(tb, db.Ping())
	return db, databaseConnectionString
}

func getDatabaseConnectionString(tb testing.TB) string {
	databaseURL, isSet := os.LookupEnv("TEST_DATABASE_URL_POSTGRES")

	if !isSet {
		tb.Skip(`"TEST_DATABASE_URL_POSTGRES" env var is not set, therefore skipping this test`)
	}

	return databaseURL
}

func TestPostgres_Close(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	s := testcase.NewSpec(t)

	pg := func(t *testcase.T) *storages.Postgres {
		return &storages.Postgres{DB: t.I(`DB`).(*sql.DB)}
	}

	subject := func(t *testcase.T) error {
		return pg(t).Close()
	}

	s.Let(`DB`, func(t *testcase.T) interface{} {
		db, _ := MustOpenDB(t)
		t.Defer(db.Close)
		return db
	})

	s.Then(`it will close the DB object`, func(t *testcase.T) {
		require.Nil(t, subject(t))

		sqlDB := t.I(`DB`).(*sql.DB)
		row := sqlDB.QueryRow(`SELECT 1=1`)
		var v sql.NullBool
		err := row.Scan(&v)
		require.Error(t, err)
		require.Contains(t, err.Error(), `closed`)
	})
}
