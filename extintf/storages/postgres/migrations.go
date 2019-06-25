package postgres

import (
	"database/sql"
	"github.com/adamluzsi/toggler/extintf/storages/postgres/assets"

	"github.com/lopezator/migrator"
)

//go:generate esc -o ./assets/fs.go -ignore fs.go -pkg assets ./assets

func (pg *Postgres) Migrate() error {

	m := migrator.New(
		&migrator.Migration{
			Name: `create Feature Flags table`,
			Func: queryToMigrationFunc(assets.FSMustString(false, `/assets/migrations/01_create_feature_flags_table.sql`)),
		},
		&migrator.Migration{
			Name: `create Pilots table`,
			Func: queryToMigrationFunc(assets.FSMustString(false, `/assets/migrations/02_create_pilots_table.sql`)),
		},
		&migrator.Migration{
			Name: `create tokens table`,
			Func: queryToMigrationFunc(assets.FSMustString(false, `/assets/migrations/03_create_tokens_table.sql`)),
		},
	)

	return m.Migrate(pg.DB)

}

func queryToMigrationFunc(q string) func(tx *sql.Tx) error {
	return func(tx *sql.Tx) error {
		if _, err := tx.Exec(q); err != nil {
			return err
		}
		return nil
	}
}
