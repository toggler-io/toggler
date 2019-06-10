package usecases_test

import (
	"github.com/adamluzsi/toggler/services/rollouts"
	. "github.com/adamluzsi/toggler/testing"
	"github.com/adamluzsi/toggler/usecases"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
)

func SetupSpec(s *testcase.Spec) {
	s.Let(`UseCases`, func(t *testcase.T) interface{} {
		return usecases.NewUseCases(GetStorage(t))
	})
}

func GetRolloutManager(t *testcase.T) *rollouts.RolloutManager {
	rm := rollouts.NewRolloutManager(GetStorage(t))
	return rm
}

func GetUseCases(t *testcase.T) *usecases.UseCases {
	return t.I(`UseCases`).(*usecases.UseCases)
}

func GetProtectedUsecases(t *testcase.T) *usecases.ProtectedUsecases {
	pu, err := GetUseCases(t).ProtectedUsecases(CreateToken(t, strconv.Itoa(rand.Int())).Token)
	require.Nil(t, err)
	require.NotNil(t, pu)
	return pu
}
