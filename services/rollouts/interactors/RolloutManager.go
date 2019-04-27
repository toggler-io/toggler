package interactors

import (
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

func (manager *RolloutManager) EnableFeatureFor(featureFlagName, ExternalPilotID string) error {

	ff, err := manager.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return err
	}

	if ff == nil {

		ff = manager.newDefaultFeatureFlag(featureFlagName)

		if err := manager.Storage.Save(ff); err != nil {
			return err
		}

	}

	pilot, err := manager.Storage.FindFlagPilotByExternalPilotID(ff.ID, ExternalPilotID)

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

	return manager.Save(&rollouts.Pilot{FeatureFlagID: ff.ID, ExternalID: ExternalPilotID, Enrolled: true})

}

func (manager *RolloutManager) newDefaultFeatureFlag(featureFlagName string) *rollouts.FeatureFlag {
	return &rollouts.FeatureFlag{
		Name: featureFlagName,
		Rollout: rollouts.Rollout{
			Percentage: 0,
			RandSeedSalt: manager.RandSeedGenerator(),
		},
	}
}
