package deployment

import (
	"context"

	"github.com/adamluzsi/frameless/resources"
)

type Storage interface {
	resources.Creator
	resources.Finder
	resources.Updater
	resources.Deleter
	resources.CreatorPublisher
	resources.UpdaterPublisher
	resources.DeleterPublisher
	resources.OnePhaseCommitProtocol
	EnvironmentFinder
}

type EnvironmentFinder interface {
	FindDeploymentEnvironmentByAlias(ctx context.Context, idOrName string, env *Environment) (bool, error)
}
