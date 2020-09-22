package caches

import (
	"time"

	"github.com/toggler-io/toggler/domains/toggler"
)

type Interface interface {
	toggler.Storage
	SetTimeToLiveForValuesToCache(duration time.Duration) error // SetTTL()
	//resources.CreatorPublisher
	//resources.UpdaterPublisher
	//resources.DeleterPublisher
}
