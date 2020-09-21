package specs

import (
	"testing"

	frmlspecs "github.com/adamluzsi/frameless/resources/specs"
	rollspecs "github.com/toggler-io/toggler/domains/release/specs"
	secuspecs "github.com/toggler-io/toggler/domains/security/specs"
	deplspecs "github.com/toggler-io/toggler/domains/deployment/specs"

	"github.com/toggler-io/toggler/usecases"
)

type Storage struct {
	Subject usecases.Storage
	frmlspecs.FixtureFactory
}

func (spec Storage) Benchmark(b *testing.B) {
	b.Run(`toggler`, func(b *testing.B) {
		rollspecs.Storage{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Benchmark(b)
		secuspecs.Storage{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Benchmark(b)
		deplspecs.Storage{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Benchmark(b)
	})
}

func (spec Storage) Test(t *testing.T) {
	t.Run(`toggler`, func(t *testing.T) {
		rollspecs.Storage{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Test(t)
		secuspecs.Storage{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Test(t)
		deplspecs.Storage{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Test(t)
	})
}
