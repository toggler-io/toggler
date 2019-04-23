package interactors

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
)

type FeatureFlagChecker struct {
	Storage rollouts.Storage
}

func (checker *FeatureFlagChecker) IsFeatureEnabledFor(featureFlagName string, ExternalPilotID string) (bool, error) {

	ff, err := checker.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return false, err
	}

	if ff == nil {
		return false, nil
	}

	if ff.Rollout.GloballyEnabled {
		return true, nil
	}

	pilot, err := checker.Storage.FindFlagPilotByExternalPilotID(ff.ID, ExternalPilotID)

	if err != nil {
		return false, err
	}

	if pilot == nil {
		return false, nil
	}

	return pilot.Enrolled, nil
}

func (checker *FeatureFlagChecker) IsFeatureGloballyEnabled(featureFlagName string) (bool, error) {
	ff, err := checker.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return false, err
	}

	if ff == nil {
		return false, nil
	}

	return ff.Rollout.GloballyEnabled, nil
}
