package storages

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/toggler-io/toggler/domains/toggler"
)

func New(connstr string) (toggler.Storage, error) {
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
		return nil, fmt.Errorf(`ErrNotImplemented`)
	}
}
