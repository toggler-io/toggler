package contracts

import (
	"github.com/adamluzsi/frameless"
	"testing"

	"github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/domains/security"
)

type Storage struct {
	Subject        func(testing.TB) security.Storage
	FixtureFactory func(testing.TB) contracts.FixtureFactory
}

func (spec Storage) String() string {
	return `security#Storage`
}

func (spec Storage) Test(t *testing.T) {
	spec.Spec(testcase.NewSpec(t))
}

func (spec Storage) Benchmark(b *testing.B) {
	spec.Spec(testcase.NewSpec(b))
}

func (spec Storage) Spec(s *testcase.Spec) {
	testcase.RunContract(s,
		TokenStorage{
			Subject: func(tb testing.TB) security.TokenStorage {
				return spec.Subject(tb).SecurityToken(spec.FixtureFactory(tb).Context())
			},
			FixtureFactory: spec.FixtureFactory,
		},
		contracts.OnePhaseCommitProtocol{T: security.Token{},
			Subject: func(tb testing.TB) (frameless.OnePhaseCommitProtocol, contracts.CRD) {
				storage := spec.Subject(tb)
				return storage, storage.SecurityToken(spec.FixtureFactory(tb).Context())
			},
			FixtureFactory: spec.FixtureFactory,
		},
	)
}
