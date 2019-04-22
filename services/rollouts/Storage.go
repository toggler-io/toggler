package rollouts

import (
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/resources/specs"
)

type Storage interface {
	specs.Save
	specs.FindByID
	specs.Truncate
	specs.DeleteByID

	FlagFinder
	PilotFinder
}

type FlagFinder interface {
	FindByFlagName(name string) (*FeatureFlag, error)
}

type PilotFinder interface {
	FindPilotByFeatureFlagIDAndPublicPilotID(FeatureFlagID, ExternalPublicPilotID string) (*Pilot, error)
	FindPilotsByFeatureFlag(*FeatureFlag) frameless.Iterator
}
