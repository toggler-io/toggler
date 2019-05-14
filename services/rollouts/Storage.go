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
	specs.Update
	specs.FindAll

	FlagFinder
	PilotFinder
}

type FlagFinder interface {
	FindByFlagName(name string) (*FeatureFlag, error)
}

type PilotFinder interface {
	FindFlagPilotByExternalPilotID(FeatureFlagID, ExternalPilotID string) (*Pilot, error)
	FindPilotsByFeatureFlag(*FeatureFlag) frameless.Iterator
}
