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
	EnvironmentFinder
}

type EnvironmentFinder interface {
	FindDeploymentEnvironmentByAlias(ctx context.Context, idOrName string, env *Environment) (bool, error)
}
