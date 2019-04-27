package usecases

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
)

func NewIsFeatureEnabledForPilotChecker(s rollouts.Storage) *IsFeatureEnabledForPilotChecker {
	return &IsFeatureEnabledForPilotChecker{
		featureFlagChecker: interactors.NewFeatureFlagChecker(s),
	}
}

type IsFeatureEnabledForPilotChecker struct {
	featureFlagChecker *interactors.FeatureFlagChecker
}

func (checker *IsFeatureEnabledForPilotChecker) IsFeatureEnabledForPilot(featureFlagName string, externalPilotID string) (bool, error) {
	return checker.featureFlagChecker.IsFeatureEnabledFor(featureFlagName, externalPilotID)
}
