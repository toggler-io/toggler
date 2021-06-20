package deployment

import (
	"context"

	"github.com/adamluzsi/frameless"
)

type Storage interface {
	frameless.OnePhaseCommitProtocol
	DeploymentEnvironment(context.Context) EnvironmentStorage
}

type EnvironmentStorage /* Environment */ interface {
	frameless.Creator
	frameless.Finder
	frameless.Updater
	frameless.Deleter
	frameless.Publisher
	EnvironmentFinder
}

type EnvironmentFinder interface {
	FindDeploymentEnvironmentByAlias(ctx context.Context, idOrName string, env *Environment) (bool, error)
}
