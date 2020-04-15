package storages

import (
	"database/sql"
	"net/url"

	"github.com/adamluzsi/frameless"

	"github.com/toggler-io/toggler/usecases"
)

func New(connstr string) (usecases.Storage, error) {
	var driver string = connstr

	u, err := url.Parse(connstr)
	if err != nil {
		return nil, err
	}

	if u.Scheme != `` {
		driver = u.Scheme
	}

	switch driver {
	case `postgres`:
		db, err := sql.Open(u.Scheme, connstr)
		if err != nil {
			return nil, err
		}
		return NewPostgres(db)

	case `memory`:
		return NewInMemory(), nil

	default:
		return nil, frameless.ErrNotImplemented
	}
}
