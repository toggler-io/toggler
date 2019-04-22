package interactors

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/frameless"
)

func IsFeatureEnabled(storage Storage, featureFlagName string, ExternalPublicPilotID string) (bool, error) {
	ff := &rollouts.FeatureFlag{}

	ok, err := storage.FindByFlagName(featureFlagName, ff)

	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	if ff.IsGlobal {
		return true, nil
	}

	pilot, err := storage.FindPilotByFeatureFlagIDAndPublicPilotID(ff.ID, ExternalPublicPilotID)

	if err != nil {
		return false, err
	}

	if pilot == nil {
		return false, nil
	}

	return true, nil
}

type FindByFlagName interface {
	FindByFlagName(name string, ptr *rollouts.FeatureFlag) (bool, error)
}

type PilotFinder interface {
	FindPilotByFeatureFlagIDAndPublicPilotID(FeatureFlagID, ExternalPublicPilotID string) (*rollouts.Pilot, error)
	FindPilotsByFeatureFlag(*rollouts.FeatureFlag) frameless.Iterator
}
