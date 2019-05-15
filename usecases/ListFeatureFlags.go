package usecases

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
)

func (uc *UseCases) ListFeatureFlags(token string) ([]*rollouts.FeatureFlag, error) {

	valid, err := uc.Doorkeeper.VerifyTokenString(token)

	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, ErrInvalidToken
	}

	return uc.RolloutManager.ListFeatureFlags()

}
