package storages

import (
	"database/sql"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/toggler/extintf/storages/postgres"
	"github.com/adamluzsi/toggler/usecases"
	"net/url"
)

func New(connstr string) (usecases.Storage, error) {
	u, err := url.Parse(connstr)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case `postgres`:
		db, err := sql.Open(u.Scheme, connstr)
		if err != nil {
			return nil, err
		}
		return postgres.NewPostgres(db)

	default:
		return nil, frameless.ErrNotImplemented
	}
}
