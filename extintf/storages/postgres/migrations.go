package postgres

import (
	"github.com/adamluzsi/toggler/extintf/storages/postgres/assets"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	_ "github.com/lib/pq"
)

//go:generate esc -o ./assets/fs.go -pkg assets -prefix assets/migrations ./assets/migrations
const migrationsDirectory = `/assets/migrations`

func (pg *Postgres) Migrate() error {

	f, err := assets.FS(false).Open(migrationsDirectory)

	if err != nil {
		return err
	}

	fis, err := f.Readdir(-1)

	if err != nil {
		return err
	}

	var names []string

	for _, fi := range fis {
		if !fi.IsDir() {
			names = append(names, fi.Name())
		}
	}

	s := bindata.Resource(names, func(name string) ([]byte, error) {
		return assets.FSByte(false, `/`+name)
	})

	srcDriver, err := bindata.WithInstance(s)

	if err != nil {
		return err
	}

	dbDriver, err := postgres.WithInstance(pg.DB, &postgres.Config{})

	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance(`esc`, srcDriver, `postgres`, dbDriver)

	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil

}
