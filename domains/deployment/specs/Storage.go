package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/resources/specs"

	"github.com/toggler-io/toggler/domains/deployment"
)

type Storage struct {
	Subject deployment.Storage
	specs.FixtureFactory
}

func (spec Storage) Test(t *testing.T) {
	t.Run(`deployment`, func(t *testing.T) {
		specs.CRUD{
			Subject:        spec.Subject,
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
		}.Test(t)

		EnvironmentFinderStorage{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Test(t)
	})
}

func (spec Storage) Benchmark(b *testing.B) {
	b.Run(`deployment`, func(b *testing.B) {
		specs.CRUD{
			Subject:        spec.Subject,
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
		}.Benchmark(b)

		EnvironmentFinderStorage{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Benchmark(b)
	})
}
