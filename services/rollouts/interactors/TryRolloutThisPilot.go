package interactors

import (
	"math/rand"
	"time"
)

func NewRolloutTrier(s Storage) *RolloutTrier {
	return &RolloutTrier{
		Storage: s,
		RandIntn: rand.New(rand.NewSource(time.Now().Unix())).Intn,
	}
}

type RolloutTrier struct {
	Storage
	RandIntn func(int) int
}

func (trier *RolloutTrier) TryRolloutThisPilot(featureFlagName string, ExternalPublicPilotID string) error {


	return nil
}
