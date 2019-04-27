package interactors

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"hash/fnv"
	"math/rand"
)

func NewFeatureFlagChecker(s rollouts.Storage) *FeatureFlagChecker {
	return &FeatureFlagChecker{
		Storage:                s,
		IDPercentageCalculator: PseudoRandPercentageWithFNV1a64,
	}
}

type FeatureFlagChecker struct {
	Storage                rollouts.Storage
	IDPercentageCalculator func(string) int
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

	return checker.IDPercentageCalculator(ExternalPilotID) <= ff.Rollout.Percentage, nil
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

func PseudoRandPercentageWithFNV1a64(id string) int {
	h := fnv.New64a()

	if _, err := h.Write([]byte(id)); err != nil {
		panic(err)
	}

	seed := rand.NewSource(int64(h.Sum64()))
	return rand.New(seed).Intn(101)
}
