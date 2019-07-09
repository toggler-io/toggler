package postgres_test

import (
	"database/sql"
	"github.com/adamluzsi/toggler/extintf/storages/postgres"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/toggler/usecases/specs"
)

func TestPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	db := MustOpenDB(t)
	defer db.Close()

	storage, err := postgres.NewPostgres(db)
	require.Nil(t, err)

	(&specs.StorageSpec{Subject: storage}).Test(t)
}

func MustOpenDB(t *testing.T) *sql.DB {
	databaseConnectionString := getDatabaseConnectionString(t)
	db, err := sql.Open("postgres", databaseConnectionString)
	require.Nil(t, err)
	require.Nil(t, db.Ping())
	return db
}

func getDatabaseConnectionString(t *testing.T) string {
	databaseURL, isSet := os.LookupEnv("TEST_STORAGE_URL_POSTGRES")

	if !isSet {
		t.Skip(`"TEST_STORAGE_URL_POSTGRES" env var is not set, therefore skipping this test`)
	}

	return databaseURL
}
