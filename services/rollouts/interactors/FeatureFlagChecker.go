package interactors

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
)

func NewFeatureFlagChecker(s rollouts.Storage) *FeatureFlagChecker {
	return &FeatureFlagChecker{
		Storage:                s,
		IDPercentageCalculator: GeneratePseudoRandPercentageWithFNV1a64,
	}
}

type FeatureFlagChecker struct {
	Storage                rollouts.Storage
	IDPercentageCalculator func(string, int64) (int, error)
}

func (checker *FeatureFlagChecker) IsFeatureEnabledFor(featureFlagName string, ExternalPilotID string) (bool, error) {

	ff, err := checker.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return false, err
	}

	if ff == nil {
		return false, nil
	}

	pilot, err := checker.Storage.FindFlagPilotByExternalPilotID(ff.ID, ExternalPilotID)

	if err != nil {
		return false, err
	}

	if pilot != nil {
		return pilot.Enrolled, nil
	}

	if ff.Rollout.Percentage == 0 {
		return false, nil
	}

	diceRollResultPercentage, err := checker.IDPercentageCalculator(ExternalPilotID, ff.Rollout.RandSeedSalt)

	if err != nil {
		return false, err
	}

	return diceRollResultPercentage <= ff.Rollout.Percentage, nil

}

func (checker *FeatureFlagChecker) IsFeatureGloballyEnabled(featureFlagName string) (bool, error) {
	ff, err := checker.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return false, err
	}

	if ff == nil {
		return false, nil
	}

	return ff.Rollout.Percentage == 100, nil
}
