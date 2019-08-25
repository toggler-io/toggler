package usecases

import (
	"github.com/adamluzsi/frameless/resources"
	"github.com/toggler-io/toggler/services/rollouts"
	"github.com/toggler-io/toggler/services/security"
	"io"
)

type Storage interface {
	resources.Saver
	resources.Finder
	resources.Updater
	resources.Deleter
	resources.Truncater
	rollouts.FlagFinder
	rollouts.PilotFinder
	security.TokenFinder
	io.Closer
}
