package contracts

import (
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/contracts"
	"testing"

	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/domains/deployment"

	sh "github.com/toggler-io/toggler/spechelper"
)

type Storage struct {
	Subject        func(testing.TB) deployment.Storage
	FixtureFactory sh.FixtureFactory
}

func (spec Storage) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec Storage) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec Storage) Spec(tb testing.TB) {
	testcase.NewSpec(tb).Describe(`deployment#Storage`, func(s *testcase.Spec) {
		testcase.RunContract(s,
			EnvironmentStorage{
				Subject: func(tb testing.TB) EnvironmentStorageSubject {
					return spec.Subject(tb).DeploymentEnvironment(spec.FixtureFactory.Context())
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.OnePhaseCommitProtocol{T: deployment.Environment{},
				Subject: func(tb testing.TB) (frameless.OnePhaseCommitProtocol, contracts.CRD) {
					storage := spec.Subject(tb)
					return storage, storage.DeploymentEnvironment(spec.FixtureFactory.Context())
				},
				FixtureFactory: spec.FixtureFactory,
			},
		)
	})
}
