package interactors

import (
	"math/rand"
	"time"
)

func NewRolloutManager(s Storage) *RolloutManager {
	return &RolloutManager{
		Storage: s,
		RandIntn: rand.New(rand.NewSource(time.Now().Unix())).Intn,
	}
}

type RolloutManager struct {
	Storage
	RandIntn func(int) int
}

func (trier *RolloutManager) TryRolloutThisPilot(featureFlagName string, ExternalPublicPilotID string) error {


	return nil
}
