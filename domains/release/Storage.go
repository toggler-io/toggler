package release

import (
	"context"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources"

	"github.com/toggler-io/toggler/domains/deployment"
)

type Storage interface {
	resources.Creator
	resources.Finder
	resources.Updater
	resources.Deleter
	resources.OnePhaseCommitProtocol
	resources.CreatorPublisher
	resources.UpdaterPublisher
	resources.DeleterPublisher
	FlagFinder
	PilotFinder
	RolloutFinder
}

type (
	PilotEntries   = iterators.Interface
	FlagEntries    = iterators.Interface
	RolloutEntries = iterators.Interface
)

type FlagFinder interface {
	FindReleaseFlagByName(ctx context.Context, name string) (*Flag, error)
	FindReleaseFlagsByName(ctx context.Context, names ...string) FlagEntries
}

type PilotFinder interface {
	FindReleaseManualPilotByExternalID(ctx context.Context, flagID, envID interface{}, pilotExtID string) (*ManualPilot, error)
	// deployment.Environment independent queries
	FindReleasePilotsByReleaseFlag(ctx context.Context, flag Flag) PilotEntries
	FindReleasePilotsByExternalID(ctx context.Context, externalID string) PilotEntries
}

type RolloutFinder interface {
	FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(context.Context, Flag, deployment.Environment, *Rollout) (bool, error)

	// TODO:
	//FindReleaseRolloutsByDeploymentEnvironment(context.Context, deployment.Environment, *Rollout) (bool, error)
}
