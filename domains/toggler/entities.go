// package usecases serves as a documentation purpose for the reader of the project.
// When the reader start examine the services, the reader can have a quick grasp on the situation,
// by simply listing the files in the usecases pkg.
package toggler

import (
	"github.com/adamluzsi/frameless"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
)

func NewUseCases(s Storage) *UseCases {
	return &UseCases{
		Storage:        s,
		RolloutManager: release.NewRolloutManager(s),
		Doorkeeper:     security.NewDoorkeeper(s),
		Issuer:         security.NewIssuer(s),
	}
}

type UseCases struct {
	Storage Storage
	*release.RolloutManager
	*security.Doorkeeper
	*security.Issuer
}

const ErrInvalidToken frameless.Error = `invalid token error`
