// package usecases serves as a documentation purpose for the reader of the project.
// When the reader start examine the services, the reader can have a quick grasp on the situation,
// by simply listing the files in the usecases pkg.
package usecases

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
)

func NewUseCases(s rollouts.Storage) *UseCases {
	return &UseCases{
		RolloutManager:     rollouts.NewRolloutManager(s),
		FeatureFlagChecker: rollouts.NewFeatureFlagChecker(s),
	}
}

type UseCases struct {
	*rollouts.FeatureFlagChecker
	*rollouts.RolloutManager
}
