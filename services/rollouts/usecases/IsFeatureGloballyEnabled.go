package usecases

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/rollouts/interactors"
)

func NewIsFeatureGloballyEnabledChecker(s rollouts.Storage) *IsFeatureGloballyEnabledChecker {
	return &IsFeatureGloballyEnabledChecker{
		FeatureFlagChecker: interactors.NewFeatureFlagChecker(s),
	}
}

type IsFeatureGloballyEnabledChecker struct {
	FeatureFlagChecker *interactors.FeatureFlagChecker
}

func (checker *IsFeatureGloballyEnabledChecker) IsFeatureGloballyEnabled(featureFlagName string) (bool, error) {
	return checker.FeatureFlagChecker.IsFeatureGloballyEnabled(featureFlagName)
}
