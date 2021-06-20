package toggler

import (
	"io"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
)

type Storage interface {
	release.Storage
	deployment.Storage
	security.Storage
	io.Closer
}
