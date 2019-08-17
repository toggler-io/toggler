package usecases

import (
	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
	"io"
)

type Storage interface {
	resources.Save
	resources.FindByID
	resources.Truncate
	resources.DeleteByID
	resources.Update
	resources.FindAll
	rollouts.FlagFinder
	rollouts.PilotFinder
	security.TokenFinder
	io.Closer
}
