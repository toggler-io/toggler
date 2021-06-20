package storages

import (
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
		return NewPostgres(connstr)

	case `memory`:
		return NewEventLogMemoryStorage(), nil

	default:
		return nil, fmt.Errorf(`ErrNotImplemented`)
	}
}
