package usecases

func (uc *UseCases) SetPilotEnrollmentForFeature(featureFlagName string, pilotExtID string, isEnrolled bool) error {
	return uc.RolloutManager.SetPilotEnrollmentForFeature(featureFlagName, pilotExtID, isEnrolled)
}
