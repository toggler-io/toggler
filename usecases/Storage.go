package usecases

import (
	"io"

	"github.com/adamluzsi/frameless/resources"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
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
