package caches

import (
	"time"

	"github.com/toggler-io/toggler/domains/toggler"
)

func NewNullCache(s toggler.Storage) *NullCache {
	return &NullCache{Storage: s}
}

type NullCache struct{ toggler.Storage }

func (*NullCache) SetTimeToLiveForValuesToCache(duration time.Duration) error {
	return nil
}
