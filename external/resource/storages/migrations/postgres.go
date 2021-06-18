package migrations

import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	pgmigr "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	_ "github.com/lib/pq"
)

func MigratePostgres(dsn string) error {
	db, err := sql.Open(`postgres`, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	m, err := NewPostgresMigrate(db)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func NewPostgresMigrate(db *sql.DB) (*migrate.Migrate, error) {

	srcDriver, err := NewPostgresBindataSourceDriver()

	if err != nil {
		return nil, err
	}

	dbDriver, err := pgmigr.WithInstance(db, &pgmigr.Config{})

	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance(`embed`, srcDriver, `postgres`, dbDriver)

	if err != nil {
		return nil, err
	}

	return m, err

}

//go:embed postgres/*
var fs embed.FS

func NewPostgresBindataSourceDriver() (source.Driver, error) {
	const dirName = `postgres`
	dirEntries, err := fs.ReadDir(dirName)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			names = append(names, dirEntry.Name())
		}
	}

	s := bindata.Resource(names, func(name string) ([]byte, error) {
		return fs.ReadFile(dirName + "/" + name)
	})

	return bindata.WithInstance(s)
}
