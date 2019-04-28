package usecases

func (uc *UseCases) UpdateFeatureFlagRolloutPercentage(featureFlagName string, rolloutPercentage int) error {
	return uc.RolloutManager.UpdateFeatureFlagRolloutPercentage(featureFlagName, rolloutPercentage)
}