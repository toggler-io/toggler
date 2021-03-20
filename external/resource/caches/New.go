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

	var cache Storage
	switch driver {
	case `redis`:
		s, err := NewRedisCacheStorage(connstr)
		if err != nil {
			return nil, err
		}
		cache = s
	case `memory`:
		cache = NewInMemoryCacheStorage()
	default:
		return src, nil
	}

	return NewManager(src, cache)
}
