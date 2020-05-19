package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/resources/specs"

	"github.com/toggler-io/toggler/domains/deployment"
)

type StorageSpec struct {
	Subject deployment.Storage
	specs.FixtureFactory
}

func (spec StorageSpec) Test(t *testing.T) {
	t.Run(`deployment`, func(t *testing.T) {
		specs.CommonSpec{
			Subject:        spec.Subject,
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
		}.Test(t)

		EnvironmentFinderStorageSpec{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Test(t)
	})
}

func (spec StorageSpec) Benchmark(b *testing.B) {
	b.Run(`deployment`, func(b *testing.B) {
		specs.CommonSpec{
			Subject:        spec.Subject,
			EntityType:     deployment.Environment{},
			FixtureFactory: spec.FixtureFactory,
		}.Benchmark(b)

		EnvironmentFinderStorageSpec{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Benchmark(b)
	})
}
