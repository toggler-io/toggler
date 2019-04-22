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

func (trier *RolloutManager) TryRolloutThisPilot(featureFlagName string, ExternalPublicPilotID string) error {


	return nil
}
