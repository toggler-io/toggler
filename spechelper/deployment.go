package spechelper

import (
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"github.com/toggler-io/toggler/domains/release"
)

const LetVarExampleDeploymentEnvironment = `example deployment environment`

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		GivenWeHaveDeploymentEnvironment(s, LetVarExampleDeploymentEnvironment)
	})
}

func GetDeploymentEnvironment(t *testcase.T, vn string) *release.Environment {
	return t.I(vn).(*release.Environment)
}

func ExampleDeploymentEnvironment(t *testcase.T) *release.Environment {
	return GetDeploymentEnvironment(t, LetVarExampleDeploymentEnvironment)
}

func GivenWeHaveDeploymentEnvironment(s *testcase.Spec, vn string) testcase.Var {
	return s.Let(vn, func(t *testcase.T) interface{} {
		de := NewFixtureFactory(t).Create(release.Environment{}).(release.Environment)
		storage := StorageGet(t).ReleaseEnvironment(ContextGet(t))
		require.Nil(t, storage.Create(ContextGet(t), &de))
		t.Defer(storage.DeleteByID, ContextGet(t), de.ID)
		t.Logf(`%#v`, de)
		return &de
	})
}

func NoDeploymentEnvironmentPresentInTheStorage(s *testcase.Spec) {
	s.Before(func(t *testcase.T) {
		require.Nil(t, StorageGet(t).ReleaseEnvironment(ContextGet(t)).DeleteAll(ContextGet(t)))
	})
}
