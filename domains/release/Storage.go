package release

import (
	"context"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"

	"github.com/toggler-io/toggler/domains/deployment"
)

type Storage interface {
	frameless.OnePhaseCommitProtocol
	ReleaseFlag(context.Context) FlagStorage
	ReleasePilot(context.Context) PilotStorage
	ReleaseRollout(context.Context) RolloutStorage
}

type (
	PilotEntries   = iterators.Interface
	FlagEntries    = iterators.Interface
	RolloutEntries = iterators.Interface
)

type FlagStorage interface {
	frameless.Creator
	frameless.Finder
	frameless.Updater
	frameless.Deleter
	frameless.Publisher
	FindReleaseFlagByName(ctx context.Context, name string) (*Flag, error)
	FindReleaseFlagsByName(ctx context.Context, names ...string) FlagEntries
}

type PilotStorage interface {
	frameless.Creator
	frameless.Finder
	frameless.Updater
	frameless.Deleter
	frameless.Publisher
	FindReleaseManualPilotByExternalID(ctx context.Context, flagID, envID interface{}, pilotExtID string) (*ManualPilot, error)
	FindReleasePilotsByReleaseFlag(ctx context.Context, flag Flag) PilotEntries
	FindReleasePilotsByExternalID(ctx context.Context, externalID string) PilotEntries
}

type RolloutStorage interface {
	frameless.Creator
	frameless.Finder
	frameless.Updater
	frameless.Deleter
	frameless.Publisher
	FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(context.Context, Flag, deployment.Environment, *Rollout) (bool, error)

	// TODO:
	//FindReleaseRolloutsByDeploymentEnvironment(context.Context, deployment.Environment, *Rollout) (bool, error)
}
