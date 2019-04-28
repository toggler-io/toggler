package interactors

import (
	"fmt"
	"time"

	"github.com/adamluzsi/FeatureFlags/services/rollouts"
)

func NewRolloutManager(s rollouts.Storage) *RolloutManager {
	return &RolloutManager{
		Storage:           s,
		RandSeedGenerator: func() int64 { return time.Now().Unix() },
	}
}

type RolloutManager struct {
	rollouts.Storage
	RandSeedGenerator func() int64
}

func (manager *RolloutManager) EnableFeatureFor(featureFlagName, externalPilotID string) error {

	ff, err := manager.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return err
	}

	if ff == nil {

		ff = manager.newDefaultFeatureFlag(featureFlagName)

		if serr := manager.Storage.Save(ff); serr != nil {
			return serr
		}

	}

	pilot, err := manager.Storage.FindFlagPilotByExternalPilotID(ff.ID, externalPilotID)

	if err != nil {
		return err
	}

	if pilot != nil {

		if pilot.Enrolled {
			return nil
		}

		if err := manager.DeleteByID(pilot, pilot.ID); err != nil {
			return err
		}

	}

	return manager.Save(&rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: externalPilotID, Enrolled: true})

}

func (manager *RolloutManager) newDefaultFeatureFlag(featureFlagName string) *rollouts.FeatureFlag {
	return &rollouts.FeatureFlag{
		Name: featureFlagName,
		Rollout: rollouts.Rollout{
			RandSeedSalt: manager.RandSeedGenerator(),
			Strategy: rollouts.RolloutStrategy{
				Percentage: 0,
			},
		},
	}
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
