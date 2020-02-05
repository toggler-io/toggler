package usecases

import (
	"github.com/adamluzsi/frameless/resources"
	"github.com/toggler-io/toggler/services/release"
	"github.com/toggler-io/toggler/services/security"
	"io"
)

type Storage interface {
	resources.Creator
	resources.Finder
	resources.Updater
	resources.Deleter
	resources.Truncater
	release.FlagFinder
	release.PilotFinder
	release.AllowFinder
	security.TokenFinder
	io.Closer
}
