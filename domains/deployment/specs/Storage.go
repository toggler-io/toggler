package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/resources/specs"

	"github.com/toggler-io/toggler/domains/deployment"

	. "github.com/toggler-io/toggler/testing"
)

type Storage struct {
	Subject        deployment.Storage
	FixtureFactory FixtureFactory
}

func (spec Storage) Test(t *testing.T) {
	t.Run(`deployment`, func(t *testing.T) {
		spec.Spec(t)
	})
}

func (spec Storage) Benchmark(b *testing.B) {
	b.Run(`deployment`, func(b *testing.B) {
		spec.Spec(b)
	})
}

func (spec Storage) Spec(tb testing.TB) {
	specs.Run(tb,
		EnvironmentStorage{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		},
	)
}
