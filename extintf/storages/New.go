package storages

import (
	"database/sql"
	"github.com/adamluzsi/toggler/extintf/storages/redis"
	"net/url"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/toggler/extintf/storages/inmemory"
	"github.com/adamluzsi/toggler/extintf/storages/postgres"
	"github.com/adamluzsi/toggler/usecases"
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
		return postgres.NewPostgres(db)

	case `redis`:
		return redis.New(connstr)

	case `memory`:
		return inmemory.New(), nil

	default:
		return nil, frameless.ErrNotImplemented
	}
}
