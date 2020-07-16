package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/resources/specs"

	"github.com/toggler-io/toggler/domains/security"
)

type StorageSpec struct {
	Subject security.Storage
	specs.FixtureFactory
}

func (spec StorageSpec) Test(t *testing.T) {
	t.Run(`security`, func(t *testing.T) {
		TokenStorageSpec{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Test(t)
	})
}

func (spec StorageSpec) Benchmark(b *testing.B) {
	b.Run(`security`, func(b *testing.B) {
		TokenStorageSpec{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Benchmark(b)
	})
}
