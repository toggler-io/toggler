package specs

import (
	"testing"

	frmlspecs "github.com/adamluzsi/frameless/resources/specs"
	rollspecs "github.com/adamluzsi/toggler/services/rollouts/specs"
	secuspecs "github.com/adamluzsi/toggler/services/security/specs"

	"github.com/adamluzsi/toggler/usecases"
)

type StorageSpec struct {
	Subject usecases.Storage

	FixtureFactory interface {
		frmlspecs.FixtureFactory
		SetPilotFeatureFlagID(ffID string) func()
	}
}

func (spec StorageSpec) Benchmark(b *testing.B) {
	b.Run(`usecases.StorageSpec`, func(b *testing.B) {
		rollspecs.StorageSpec{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Benchmark(b)
		secuspecs.StorageSpec{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Benchmark(b)
	})
}

func (spec StorageSpec) Test(t *testing.T) {
	t.Run(`usecases.StorageSpec`, func(t *testing.T) {
		rollspecs.StorageSpec{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Test(t)
		secuspecs.StorageSpec{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Test(t)
	})
}
