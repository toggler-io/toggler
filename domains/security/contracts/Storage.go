package contracts

import (
	"github.com/adamluzsi/frameless"
	"testing"

	"github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/domains/security"
)

type Storage struct {
	Subject func(testing.TB) security.Storage
	contracts.FixtureFactory
}

func (spec Storage) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec Storage) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec Storage) Spec(tb testing.TB) {
	testcase.NewSpec(tb).Describe(`security#Storage`, func(s *testcase.Spec) {
		testcase.RunContract(s,
			TokenStorage{
				Subject: func(tb testing.TB) security.TokenStorage {
					return spec.Subject(tb).SecurityToken(spec.Context())
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.OnePhaseCommitProtocol{T: security.Token{},
				Subject: func(tb testing.TB) (frameless.OnePhaseCommitProtocol, contracts.CRD) {
					storage := spec.Subject(tb)
					return storage, storage.SecurityToken(spec.Context())
				},
				FixtureFactory: spec.FixtureFactory,
			},
		)
	})
}
