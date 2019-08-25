package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/toggler-io/toggler/services/security"
)

type StorageSpec struct {
	Subject security.Storage
	specs.FixtureFactory
}

func (spec StorageSpec) Benchmark(b *testing.B) {
	b.Run(`security`, func(b *testing.B) {
		b.Run(`Token`, func(b *testing.B) {
			entityType := security.Token{}
			specs.CommonSpec{EntityType: entityType, FixtureFactory: spec.FixtureFactory, Subject: spec.Subject}.Benchmark(b)
			TokenFinderSpec{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Benchmark(b)
		})
	})
}

func (spec StorageSpec) Test(t *testing.T) {
	t.Run(`security`, func(t *testing.T) {
		t.Run(`Token`, func(t *testing.T) {
			entityType := security.Token{}
			specs.CommonSpec{EntityType: entityType, FixtureFactory: spec.FixtureFactory, Subject: spec.Subject}.Test(t)
			TokenFinderSpec{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Test(t)
		})
	})
}
