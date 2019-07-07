package caches

import (
	"github.com/adamluzsi/toggler/usecases"
	"time"
)

type Interface interface {
	usecases.Storage
	SetTimeToLiveForValuesToCache(duration time.Duration) error
}
