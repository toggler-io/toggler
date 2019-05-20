package usecases

func (uc *UseCases) UpdateFeatureFlagRolloutPercentage(token string, featureFlagName string, rolloutPercentage int) error {

	valid, err := uc.Doorkeeper.VerifyTokenString(token)

	if err != nil {
		return err
	}

	if !valid {
		return ErrInvalidToken
	}

	return uc.RolloutManager.UpdateFeatureFlagRolloutPercentage(featureFlagName, rolloutPercentage)

}
