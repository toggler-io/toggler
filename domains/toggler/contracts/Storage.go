package contracts

import (
	"context"
	"testing"

	sh "github.com/toggler-io/toggler/spechelper"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"

	relspecs "github.com/toggler-io/toggler/domains/release/contracts"
	secspecs "github.com/toggler-io/toggler/domains/security/contracts"

	"github.com/toggler-io/toggler/domains/toggler"
)

type Storage struct {
	Subject        func(testing.TB) toggler.Storage
	Context        func(testing.TB) context.Context
	FixtureFactory func(testing.TB) frameless.FixtureFactory
}

func (c Storage) Test(t *testing.T) {
	c.Spec(sh.NewSpec(t))
}

func (c Storage) Benchmark(b *testing.B) {
	c.Spec(sh.NewSpec(b))
}

func (c Storage) String() string {
	return "toggler#Storage"
}

func (c Storage) Spec(s *testcase.Spec) {
	sh.Storage.Let(s, func(t *testcase.T) interface{} {
		return c.Subject(t)
	})
	testcase.RunContract(s,
		relspecs.Storage{
			Subject: func(tb testing.TB) release.Storage {
				return c.Subject(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		secspecs.Storage{
			Subject: func(tb testing.TB) security.Storage {
				return c.Subject(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
	)
}
