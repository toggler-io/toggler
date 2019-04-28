package usecases

func (uc *UseCases) IsFeatureEnabledFor(featureFlagName string, externalPilotID string) (bool, error) {
	return uc.FeatureFlagChecker.IsFeatureEnabledFor(featureFlagName, externalPilotID)
}
