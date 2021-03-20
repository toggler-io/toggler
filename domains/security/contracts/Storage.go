package contracts

import (
	"testing"

	"github.com/adamluzsi/frameless/resources/contracts"
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
				Subject:        spec.Subject,
				FixtureFactory: spec.FixtureFactory,
			},
		)
	})
}
