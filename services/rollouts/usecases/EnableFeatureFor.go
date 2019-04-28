package usecases

func (uc *UseCases) EnableFeatureFor(featureFlagName, externalPilotID string) error {
	return uc.RolloutManager.EnableFeatureFor(featureFlagName, externalPilotID)
}
