package specs

import (
	"testing"

	"github.com/toggler-io/toggler/domains/release"
	. "github.com/toggler-io/toggler/testing"
)

type Storage struct {
	Subject        release.Storage
	FixtureFactory FixtureFactory
}

func (spec Storage) Test(t *testing.T) {
	t.Run(`releases`, func(t *testing.T) {
		RolloutStorage{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Test(t)

		FlagStorage{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Test(t)

		ManualPilotStorage{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Test(t)
	})
}

func (spec Storage) Benchmark(b *testing.B) {
	b.Run(`releases`, func(b *testing.B) {
		b.Skip()
	})
}
