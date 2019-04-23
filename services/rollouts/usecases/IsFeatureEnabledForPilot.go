package usecases

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
)

func NewIsFeatureEnabledForPilotChecker(s rollouts.Storage) *IsFeatureEnabledForPilotChecker {
	return &IsFeatureEnabledForPilotChecker{
		rolloutManager:     interactors.NewRolloutManager(s),
		featureFlagChecker: &interactors.FeatureFlagChecker{Storage: s},
	}
}

type IsFeatureEnabledForPilotChecker struct {
	rolloutManager     *interactors.RolloutManager
	featureFlagChecker *interactors.FeatureFlagChecker
}

func (checker *IsFeatureEnabledForPilotChecker) IsFeatureEnabledForPilot(featureFlagName string, ExternalPilotID string) (bool, error) {
	if err := checker.rolloutManager.TryRolloutThisPilot(featureFlagName, ExternalPilotID); err != nil {
		return false, err
	}

	return checker.featureFlagChecker.IsFeatureEnabledFor(featureFlagName, ExternalPilotID)
}
