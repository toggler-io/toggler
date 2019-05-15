package usecases

func (uc *UseCases) SetPilotEnrollmentForFeature(token, featureFlagName, pilotExtID string, isEnrolled bool) error {

	valid, err := uc.Doorkeeper.VerifyTokenString(token)

	if err != nil {
		return err
	}

	if !valid {
		return ErrInvalidToken
	}

	return uc.RolloutManager.SetPilotEnrollmentForFeature(featureFlagName, pilotExtID, isEnrolled)

}
