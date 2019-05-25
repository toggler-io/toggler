package usecases

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
)

func (uc *UseCases) SetFeatureFlag(token string, flag *rollouts.FeatureFlag) error {

	valid, err := uc.Doorkeeper.VerifyTokenString(token)

	if err != nil {
		return err
	}

	if !valid {
		return ErrInvalidToken
	}

	return uc.RolloutManager.SetFeatureFlag(flag)

}
