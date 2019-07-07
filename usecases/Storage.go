package usecases

import (
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
	"io"
)

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
	io.Closer
}
