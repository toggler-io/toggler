package interactors

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
)

func NewRolloutManager(s rollouts.Storage) *RolloutManager {
	return &RolloutManager{Storage: s}
}

type RolloutManager struct {
	rollouts.Storage
}

func (manager *RolloutManager) EnableFeatureFor(featureFlagName, ExternalPilotID string) error {

	ff, err := manager.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return err
	}

	if ff == nil {

		ff = &rollouts.FeatureFlag{Name: featureFlagName, Rollout: rollouts.Rollout{Percentage: 0}}

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
