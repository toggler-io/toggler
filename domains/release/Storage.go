package release

import (
	"context"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
)

type Storage interface {
	frameless.OnePhaseCommitProtocol
	ReleaseFlag(context.Context) FlagStorage
	ReleasePilot(context.Context) PilotStorage
	ReleaseRollout(context.Context) RolloutStorage
	ReleaseEnvironment(context.Context) EnvironmentStorage
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
	FindByName(ctx context.Context, name string) (*Flag, error)
	FindByNames(ctx context.Context, names ...string) FlagEntries
}

type PilotStorage interface {
	frameless.Creator
	frameless.Finder
	frameless.Updater
	frameless.Deleter
	frameless.Publisher
	FindByFlagEnvPublicID(ctx context.Context, flagID, envID interface{}, publicID string) (*Pilot, error)
	FindByFlag(ctx context.Context, flag Flag) PilotEntries
	FindByPublicID(ctx context.Context, publicID string) PilotEntries
}

type RolloutStorage interface {
	frameless.Creator
	frameless.Finder
	frameless.Updater
	frameless.Deleter
	frameless.Publisher
	FindByFlagEnvironment(context.Context, Flag, Environment, *Rollout) (bool, error)

	// TODO:
	//FindReleaseRolloutsByDeploymentEnvironment(context.Context, deployment.Environment, *Rollout) (bool, error)
}

type EnvironmentStorage /* Environment */ interface {
	frameless.Creator
	frameless.Finder
	frameless.Updater
	frameless.Deleter
	frameless.Publisher
	FindByAlias(ctx context.Context, idOrName string, env *Environment) (bool, error)
}
