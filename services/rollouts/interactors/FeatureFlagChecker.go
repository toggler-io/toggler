package interactors

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
)

type FeatureFlagChecker struct {
	Storage Storage
}

func (ffc *FeatureFlagChecker) IsFeatureEnabled(featureFlagName string, ExternalPublicPilotID string) (bool, error) {
	ff := &rollouts.FeatureFlag{}

	ff, err := ffc.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return false, err
	}

	if ff == nil {
		return false, nil
	}

	if ff.Rollout.GloballyEnabled {
		return true, nil
	}

	pilot, err := ffc.Storage.FindPilotByFeatureFlagIDAndPublicPilotID(ff.ID, ExternalPublicPilotID)

	if err != nil {
		return false, err
	}

	if pilot == nil {
		return false, nil
	}

	return true, nil
}
