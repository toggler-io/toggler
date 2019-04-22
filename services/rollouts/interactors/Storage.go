package interactors

import (
	"github.com/adamluzsi/frameless/resources/specs"
)

type Storage interface {
	specs.Resource
	FindByFlagName
	PilotFinder
}
