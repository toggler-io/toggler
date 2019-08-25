package caches

import (
	"net/url"

	"github.com/toggler-io/toggler/extintf/caches/nullcache"
	"github.com/toggler-io/toggler/usecases"
)

func New(connstr string, storage usecases.Storage) (Interface, error) {
	var driver string = connstr

	u, err := url.Parse(connstr)
	if err != nil {
		return nil, err
	}

	if u.Scheme != `` {
		driver = u.Scheme
	}

	switch driver {
	default:
		return nullcache.NewNullCache(storage), nil
	}
}
