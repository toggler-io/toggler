package contracts

import (
	"context"
	"testing"

	"github.com/adamluzsi/frameless"

	"github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/domains/security"
)

type Storage struct {
	Subject        func(testing.TB) security.Storage
	Context        func(testing.TB) context.Context
	FixtureFactory func(testing.TB) frameless.FixtureFactory
}

func (c Storage) String() string {
	return `security#Storage`
}

func (c Storage) Test(t *testing.T) {
	c.Spec(testcase.NewSpec(t))
}

func (c Storage) Benchmark(b *testing.B) {
	c.Spec(testcase.NewSpec(b))
}

func (c Storage) Spec(s *testcase.Spec) {
	testcase.RunContract(s,
		TokenStorage{
			Subject: func(tb testing.TB) security.TokenStorage {
				return c.Subject(tb).SecurityToken(c.Context(tb))
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		contracts.OnePhaseCommitProtocol{T: security.Token{},
			Subject: func(tb testing.TB) (frameless.OnePhaseCommitProtocol, contracts.CRD) {
				storage := c.Subject(tb)
				return storage, storage.SecurityToken(c.Context(tb))
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
	)
}
