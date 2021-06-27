package storages_test

import (
	"context"
	"database/sql"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/external/resource/storages/migrations"
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

var (
	_ toggler.Storage            = &storages.Postgres{}
	_ release.Storage            = &storages.Postgres{}
	_ security.Storage           = &storages.Postgres{}
	_ release.EnvironmentStorage = &storages.ReleaseEnvironmentPgStorage{}
	_ release.FlagStorage        = &storages.ReleaseFlagPgStorage{}
	_ release.RolloutStorage     = &storages.ReleaseRolloutPgStorage{}
	_ release.PilotStorage       = &storages.ReleasePilotPgStorage{}
)

func BenchmarkPostgres(b *testing.B) {
	if testing.Short() {
		b.Skip()
	}

	storage, err := storages.NewPostgres(getDatabaseConnectionString(b))
	require.Nil(b, err)
	defer storage.Close()

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

	storage, err := storages.NewPostgres(getDatabaseConnectionString(t))
	require.Nil(t, err)
	defer storage.Close()

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

	pg := s.Let(`*storages.Postgres`, func(t *testcase.T) interface{} {
		pg, err := storages.NewPostgres(getDatabaseConnectionString(t))
		require.Nil(t, err)
		return pg
	})
	pgGet := func(t *testcase.T) *storages.Postgres {
		return pg.Get(t).(*storages.Postgres)
	}

	subject := func(t *testcase.T) error {
		return pgGet(t).Close()
	}

	s.Then(`it will close the DB object`, func(t *testcase.T) {
		c, err := pgGet(t).ConnectionManager.GetConnection(sh.ContextGet(t))
		require.Nil(t, err)
		require.Nil(t, subject(t))
		_, err = c.ExecContext(context.Background(), `SELECT 1`)
		require.Error(t, err)
		require.Contains(t, err.Error(), `closed`)
	})
}

func TestPostgres_migration(t *testing.T) {
	require.Nil(t, migrations.MigratePostgres(getDatabaseConnectionString(t)))
}
