package usecases

import (
	"context"
	"github.com/toggler-io/toggler/services/release"
	"github.com/toggler-io/toggler/services/security"
)

type ProtectedUsecases struct {
	*release.RolloutManager
	*security.Doorkeeper
	*security.Issuer
}

func (uc *UseCases) ProtectedUsecases(ctx context.Context, token string) (*ProtectedUsecases, error) {

	valid, err := uc.protectedUsecases.Doorkeeper.VerifyTextToken(ctx, token)

	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, ErrInvalidToken
	}

	return uc.protectedUsecases, nil

}
