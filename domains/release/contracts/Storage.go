package contracts

import (
	"context"
	"testing"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/domains/release"
)

type Storage struct {
	Subject        func(testing.TB) release.Storage
	Context        func(testing.TB) context.Context
	FixtureFactory func(testing.TB) frameless.FixtureFactory
}

func (c Storage) String() string {
	return `releases#Storage`
}

func (c Storage) Test(t *testing.T) {
	c.Spec(testcase.NewSpec(t))
}

func (c Storage) Benchmark(b *testing.B) {
	c.Spec(testcase.NewSpec(b))
}

func (c Storage) Spec(s *testcase.Spec) {
	testcase.RunContract(s,
		RolloutStorage{
			Subject: func(tb testing.TB) release.Storage {
				return c.Subject(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		FlagStorage{
			Subject: func(tb testing.TB) release.Storage {
				return c.Subject(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		PilotStorage{
			Subject: func(tb testing.TB) release.Storage {
				return c.Subject(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		EnvironmentStorage{
			Subject: func(tb testing.TB) release.Storage {
				return c.Subject(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
	)
}
