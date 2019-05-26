package usecases

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/security"
)

type ProtectedUsecases struct {
	*rollouts.RolloutManager
	*security.Doorkeeper
	*security.Issuer
}

func (uc *UseCases) ProtectedUsecases(token string) (*ProtectedUsecases, error) {

	valid, err := uc.protectedUsecases.Doorkeeper.VerifyTokenString(token)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, ErrInvalidToken
	}

	return uc.protectedUsecases, nil

}
