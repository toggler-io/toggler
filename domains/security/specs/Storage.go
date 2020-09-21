package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/resources/specs"

	"github.com/toggler-io/toggler/domains/security"
)

type Storage struct {
	Subject security.Storage
	specs.FixtureFactory
}

func (spec Storage) Test(t *testing.T) {
	t.Run(`security`, func(t *testing.T) {
		TokenStorage{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Test(t)
	})
}

func (spec Storage) Benchmark(b *testing.B) {
	b.Run(`security`, func(b *testing.B) {
		TokenStorage{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Benchmark(b)
	})
}
