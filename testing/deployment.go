package testing

import (
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/deployment"
)

const ExampleDeploymentEnvironmentLetVar = `testing example deployment environment`

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(ExampleDeploymentEnvironmentLetVar, func(t *testcase.T) interface{} {
			de := FixtureFactory{}.Create(deployment.Environment{}).(*deployment.Environment)
			require.Nil(t, ExampleStorage(t).Create(GetContext(t), de))
			t.Defer(func() { _ = ExampleStorage(t).DeleteByID(GetContext(t), *de, de.ID) })
			return de
		})
	})
}

func GetDeploymentEnvironment(t *testcase.T, vn string) *deployment.Environment {
	return t.I(vn).(*deployment.Environment)
}

func ExampleDeploymentEnvironment(t *testcase.T) *deployment.Environment {
	return GetDeploymentEnvironment(t, ExampleDeploymentEnvironmentLetVar)
}
