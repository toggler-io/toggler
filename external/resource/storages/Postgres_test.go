package storages_test

import (
	"context"
	csh "github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/resource/storages/migrations"
	"os"
	"testing"

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

func TestPostgres(t *testing.T)      { SpecPostgres(t) }
func BenchmarkPostgres(b *testing.B) { SpecPostgres(b) }

func SpecPostgres(tb testing.TB) {
	if testing.Short() {
		tb.Skip()
	}

	storage, err := storages.NewPostgres(getDatabaseConnectionString(tb))
	require.Nil(tb, err)
	defer storage.Close()

	testcase.RunContract(sh.NewSpec(tb), contracts.Storage{
		Subject: func(tb testing.TB) toggler.Storage {
			return storage
		},
		FixtureFactory: func(tb testing.TB) csh.FixtureFactory {
			return sh.NewFixtureFactory(tb)
		},
	})
}

func getDatabaseConnectionString(tb testing.TB) string {
	databaseURL, isSet := os.LookupEnv("TEST_DATABASE_URL_POSTGRES")
	if !isSet {
		tb.Skip(`"TEST_DATABASE_URL_POSTGRES" env var is not set, therefore skipping this test`)
	}
	// Hack, to ensure that fixture factory creates entities in the same database as this test
	// instead of the inmemory variant.
	testcase.SetEnv(tb, `TEST_DATABASE_URL`, databaseURL)
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
