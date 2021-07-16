package contracts_test

import (
	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/domains/release/contracts"
)

var _ = []testcase.Contract{
	contracts.Storage{},
	contracts.RolloutStorage{},
	contracts.FlagStorage{},
	contracts.PilotStorage{},
	contracts.EnvironmentStorage{},
}
