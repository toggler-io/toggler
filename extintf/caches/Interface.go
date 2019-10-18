package caches

import (
	"github.com/toggler-io/toggler/usecases"
	"time"
)

type Interface interface {
	usecases.Storage
	SetTimeToLiveForValuesToCache(duration time.Duration) error // SetTTL()
}
