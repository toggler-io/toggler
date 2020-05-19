package testing

import (
	"github.com/adamluzsi/testcase"

	"github.com/toggler-io/toggler/external/resource/storages"
	"github.com/toggler-io/toggler/usecases"
)

const ExampleStorageLetVar = `TestStorage`

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(ExampleStorageLetVar, func(t *testcase.T) interface{} {
			return storages.NewInMemory()
		})
	})
}

func GetStorage(t *testcase.T, varName string) usecases.Storage {
	return t.I(varName).(usecases.Storage)
}

func ExampleStorage(t *testcase.T) usecases.Storage {
	return GetStorage(t, ExampleStorageLetVar)
}
