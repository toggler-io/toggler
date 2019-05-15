// package usecases serves as a documentation purpose for the reader of the project.
// When the reader start examine the services, the reader can have a quick grasp on the situation,
// by simply listing the files in the usecases pkg.
package usecases

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/security"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/resources/specs"
)

func NewUseCases(s Storage) *UseCases {
	return &UseCases{
		FeatureFlagChecker: rollouts.NewFeatureFlagChecker(s),
		RolloutManager:     rollouts.NewRolloutManager(s),
		Doorkeeper:         security.NewDoorkeeper(s),
		Issuer:             security.NewIssuer(s),
	}
}

type UseCases struct {
	*rollouts.FeatureFlagChecker
	*rollouts.RolloutManager
	*security.Doorkeeper
	*security.Issuer
}

type Storage interface {
	specs.Save
	specs.FindByID
	specs.Truncate
	specs.DeleteByID
	specs.Update
	specs.FindAll
	rollouts.FlagFinder
	rollouts.PilotFinder
	security.TokenFinder
}

const ErrInvalidToken frameless.Error = `invalid token error`
