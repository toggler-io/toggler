package usecases

func (uc *UseCases) IsFeatureGloballyEnabled(featureFlagName string) (bool, error) {
	return uc.FeatureFlagChecker.IsFeatureGloballyEnabled(featureFlagName)
}
