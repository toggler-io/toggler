// package usecases serves as a documentation purpose for the reader of the project.
// When the reader start examine the services, the reader can have a quick grasp on the situation,
// by simply listing the files in the usecases pkg.
package usecases

import (
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
)

func NewUseCases(s Storage) *UseCases {
	return &UseCases{
		FeatureFlagChecker: rollouts.NewFeatureFlagChecker(s),
		protectedUsecases: &ProtectedUsecases{
			RolloutManager: rollouts.NewRolloutManager(s),
			Doorkeeper:     security.NewDoorkeeper(s),
			Issuer:         security.NewIssuer(s),
		},
	}
}

type UseCases struct {
	*rollouts.FeatureFlagChecker
	protectedUsecases *ProtectedUsecases
}

const ErrInvalidToken frameless.Error = `invalid token error`
