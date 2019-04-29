package rollouts

import (
	"fmt"
	"time"
)

func NewRolloutManager(s Storage) *RolloutManager {
	return &RolloutManager{
		Storage:           s,
		RandSeedGenerator: func() int64 { return time.Now().Unix() },
	}
}

// RolloutManager provides you with feature flag configurability.
// The manager use storage in a write heavy behavior.
//
// SRP: release manager
type RolloutManager struct {
	Storage
	RandSeedGenerator func() int64
}

func (manager *RolloutManager) SetPilotEnrollmentForFeature(featureFlagName string, pilotExtID string, isEnrolled bool) error {

	ff, err := manager.ensureFeatureFlag(featureFlagName)

	if err != nil {
		return err
	}

	pilot, err := manager.Storage.FindFlagPilotByExternalPilotID(ff.ID, pilotExtID)

	if err != nil {
		return err
	}

	if pilot != nil {
		pilot.Enrolled = isEnrolled
		return manager.Storage.Update(pilot)
	}

	return manager.Save(&Pilot{FeatureFlagID: ff.ID, ExternalID: pilotExtID, Enrolled: isEnrolled})

}

func (manager *RolloutManager) UpdateFeatureFlagRolloutPercentage(featureFlagName string, rolloutPercentage int) error {

	if rolloutPercentage < 0 || 100 < rolloutPercentage {
		return fmt.Errorf(`validation error, percentage value not acceptable: %d`, rolloutPercentage)
	}

	ff, err := manager.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return err
	}

	if ff == nil {
		ff = manager.newDefaultFeatureFlag(featureFlagName)
		ff.Rollout.Strategy.Percentage = rolloutPercentage
		return manager.Storage.Save(ff)
	}

	ff.Rollout.Strategy.Percentage = rolloutPercentage
	return manager.Storage.Update(ff)

}

//----------------------------------------------------------------------------------------------------------------------

func (manager *RolloutManager) ensureFeatureFlag(featureFlagName string) (*FeatureFlag, error) {

	ff, err := manager.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return nil, err
	}

	if ff == nil {
		ff = manager.newDefaultFeatureFlag(featureFlagName)
		err = manager.Storage.Save(ff)
	}

	return ff, nil

}

func (manager *RolloutManager) newDefaultFeatureFlag(featureFlagName string) *FeatureFlag {
	return &FeatureFlag{
		Name: featureFlagName,
		Rollout: Rollout{
			RandSeedSalt: manager.RandSeedGenerator(),
			Strategy: RolloutStrategy{
				Percentage: 0,
			},
		},
	}
}
