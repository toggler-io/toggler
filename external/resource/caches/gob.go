package caches

import (
	"encoding/gob"

	"github.com/toggler-io/toggler/domains/release"
)

func init() {
	gob.Register(release.RolloutDecisionNOT{})
	gob.Register(release.RolloutDecisionAND{})
	gob.Register(release.RolloutDecisionOR{})
	gob.Register(release.RolloutDecisionByAPI{})
	gob.Register(release.RolloutDecisionByPercentage{})
	gob.Register(release.RolloutDecisionByGlobal{})
}
