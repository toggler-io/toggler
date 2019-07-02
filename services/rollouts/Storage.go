package rollouts

import (
	"context"

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
	FindFlagByName(ctx context.Context, name string) (*FeatureFlag, error)
}

type PilotFinder interface {
	FindFlagPilotByExternalPilotID(ctx context.Context, FeatureFlagID, ExternalPilotID string) (*Pilot, error)
	FindPilotsByFeatureFlag(ctx context.Context, ff *FeatureFlag) frameless.Iterator
}
