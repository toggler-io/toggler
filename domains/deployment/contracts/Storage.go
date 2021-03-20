package contracts

import (
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
					return spec.Subject(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
		)
	})
}
