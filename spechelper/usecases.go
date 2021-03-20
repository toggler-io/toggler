package spechelper

import (
	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/domains/toggler"
)

const LetVarExampleUseCases = `LetVarExampleUseCases`

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(LetVarExampleUseCases, func(t *testcase.T) interface{} {
			return toggler.NewUseCases(StorageGet(t))
		})
	})
}

func GetUseCases(t *testcase.T, varName string) *toggler.UseCases {
	return t.I(varName).(*toggler.UseCases)
}

func ExampleUseCases(t *testcase.T) *toggler.UseCases {
	return GetUseCases(t, LetVarExampleUseCases)
}
