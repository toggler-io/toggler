package usecases_test

import (
	"context"
	"math/rand"
	"strconv"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases"
)

// Deprecated
func GetProtectedUsecases(t *testcase.T) *usecases.ProtectedUseCases {
	tt, _ := CreateToken(t, strconv.Itoa(rand.Int()))
	pu, err := ExampleUseCases(t).ProtectedUsecases(context.Background(), tt)
	require.Nil(t, err)
	require.NotNil(t, pu)
	return pu
}
