package caches

import (
	"net/url"

	"github.com/toggler-io/toggler/domains/toggler"
)

func New(connstr string, src toggler.Storage) (toggler.Storage, error) {
	var driver string = connstr

	u, err := url.Parse(connstr)
	if err != nil {
		return nil, err
	}

	if u.Scheme != `` {
		driver = u.Scheme
	}

	switch driver {
	case "memory":
		return NewMemory(src)

	default:
		return src, nil
	}
}
