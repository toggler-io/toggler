package contracts

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/domains/release"
	sh "github.com/toggler-io/toggler/spechelper"
)

type Storage struct {
	Subject        func(testing.TB) release.Storage
	FixtureFactory sh.FixtureFactory
}

func (spec Storage) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec Storage) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec Storage) Spec(tb testing.TB) {
	testcase.NewSpec(tb).Describe(`releases#Storage`, func(s *testcase.Spec) {
		testcase.RunContract(s,
			RolloutStorage{
				Subject: func(tb testing.TB) release.Storage {
					return spec.Subject(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			FlagStorage{
				Subject: func(tb testing.TB) release.Storage {
					return spec.Subject(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			PilotStorage{
				Subject: func(tb testing.TB) release.Storage {
					return spec.Subject(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			EnvironmentStorage{
				Subject: func(tb testing.TB) release.Storage {
					return spec.Subject(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
		)

	})
}
