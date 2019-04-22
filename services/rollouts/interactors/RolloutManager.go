package interactors

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"math/rand"
	"time"
)

func NewRolloutManager(s rollouts.Storage) *RolloutManager {
	return &RolloutManager{
		Storage: s,
		RandIntn: rand.New(rand.NewSource(time.Now().Unix())).Intn,
	}
}

type RolloutManager struct {
	rollouts.Storage
	RandIntn func(int) int
}

func (manager *RolloutManager) TryRolloutThisPilot(featureFlagName string, ExternalPilotID string) error {

	ff, err := manager.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return err
	}

	pilot, err := manager.Storage.FindFlagPilotByExternalPilotID(ff.ID, ExternalPilotID)

	if err != nil {
		return err
	}

	if pilot != nil {
		return nil
	}

	return manager.Storage.Save(&rollouts.Pilot{
		FeatureFlagID: ff.ID,
		ExternalID: ExternalPilotID,
		Enrolled: manager.tryLuckForFeatureEnrollmentWith(ff),
	})

}

func (manager *RolloutManager) tryLuckForFeatureEnrollmentWith(ff *rollouts.FeatureFlag) bool {
	nextRand := manager.RandIntn(99) + 1
	return nextRand <= ff.Rollout.Percentage
}
