package usecases_test

import (
	"context"
	"github.com/toggler-io/toggler/services/release"
	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases"
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

func GetRolloutManager(t *testcase.T) *release.RolloutManager {
	rm := release.NewRolloutManager(GetStorage(t))
	return rm
}

func GetUseCases(t *testcase.T) *usecases.UseCases {
	return t.I(`UseCases`).(*usecases.UseCases)
}

func GetProtectedUsecases(t *testcase.T) *usecases.ProtectedUsecases {
	tt, _ := CreateToken(t, strconv.Itoa(rand.Int()))
	pu, err := GetUseCases(t).ProtectedUsecases(context.Background(), tt)
	require.Nil(t, err)
	require.NotNil(t, pu)
	return pu
}
