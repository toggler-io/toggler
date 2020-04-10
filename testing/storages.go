package testing

import (
	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/extintf/storages"
)

const ExampleStorageLetVar = `TestStorage`

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(ExampleStorageLetVar, func(t *testcase.T) interface{} {
			return storages.NewInMemory()
		})
	})
}

func GetStorage(t *testcase.T, varName string) *storages.InMemory {
	return t.I(varName).(*storages.InMemory)
}

func ExampleStorage(t *testcase.T) *storages.InMemory {
	return GetStorage(t, ExampleStorageLetVar)
}
