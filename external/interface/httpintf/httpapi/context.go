package httpapi

import (
	"context"

	"github.com/toggler-io/toggler/usecases"
)


func GetProtectedUseCases(ctx context.Context) *usecases.ProtectedUseCases {
	return ctx.Value(`ProtectedUseCases`).(*usecases.ProtectedUseCases)
}
