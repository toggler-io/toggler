package caches

import (
	"net/url"

	"github.com/toggler-io/toggler/domains/toggler"
)

func New(connstr string, storage toggler.Storage) (Interface, error) {
	var driver string = connstr

	u, err := url.Parse(connstr)
	if err != nil {
		return nil, err
	}

	if u.Scheme != `` {
		driver = u.Scheme
	}

	switch driver {
	case `redis`:
		return NewRedis(connstr, storage)
	case `memory`:
		return NewInMemory(storage), nil
	default:
		return NewNullCache(storage), nil
	}
}
