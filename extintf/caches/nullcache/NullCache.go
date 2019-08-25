package nullcache

import (
	"github.com/toggler-io/toggler/usecases"
	"time"
)

func NewNullCache(s usecases.Storage) *NullCache {
	return &NullCache{Storage: s}
}

type NullCache struct{ usecases.Storage }

func (*NullCache) SetTimeToLiveForValuesToCache(duration time.Duration) error {
	return nil
}
