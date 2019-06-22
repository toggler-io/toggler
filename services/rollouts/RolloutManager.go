package rollouts

import (
	"github.com/adamluzsi/frameless"
	"time"

	"github.com/adamluzsi/frameless/iterators"
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

func (manager *RolloutManager) CreateFeatureFlag(flag *FeatureFlag) error {
	if flag == nil {
		return ErrMissingFlag
	}

	if err := flag.Verify(); err != nil {
		return err
	}

	if flag.ID != `` {
		return ErrInvalidAction
	}

	if flag.Rollout.RandSeed == 0 {
		flag.Rollout.RandSeed = manager.RandSeedGenerator()
	}

	ff, err :=  manager.Storage.FindFlagByName(flag.Name)

	if err != nil {
		return err
	}

	if ff != nil {
		//TODO: this should be handled in transaction!
		// as mvp solution, it is acceptable for now,
		// but spec must be moved to the storage specs as `name is uniq across entries`
		return ErrFlagAlreadyExist
	}

	return manager.Storage.Save(flag)
}

func (manager *RolloutManager) UpdateFeatureFlag(flag *FeatureFlag) error {
	if flag == nil {
		return ErrMissingFlag
	}

	if  err := flag.Verify(); err != nil {
		return err
	}

	if flag.ID == `` {
		ff, err := manager.Storage.FindFlagByName(flag.Name)
		if err != nil {
			return err
		}

		if ff != nil {
			flag.ID = ff.ID

			flag.Rollout.RandSeed = ff.Rollout.RandSeed
		}
	}

	if flag.Rollout.RandSeed == 0 {
		flag.Rollout.RandSeed = manager.RandSeedGenerator()
	}

	return manager.Storage.Update(flag)
}

func (manager *RolloutManager) ListFeatureFlags() ([]*FeatureFlag, error) {
	iter := manager.Storage.FindAll(FeatureFlag{})
	ffs := []*FeatureFlag{} // empty slice required for null object pattern enforcement
	err := iterators.CollectAll(iter, &ffs)
	return ffs, err
}

func (manager *RolloutManager) SetPilotEnrollmentForFeature(flagID, pilotID string, isEnrolled bool) error {

	var ff FeatureFlag

	found, err := manager.Storage.FindByID(flagID, &ff)

	if err != nil {
		return err
	}

	if !found {
		return frameless.ErrNotFound
	}

	pilot, err := manager.Storage.FindFlagPilotByExternalPilotID(ff.ID, pilotID)

	if err != nil {
		return err
	}

	if pilot != nil {
		pilot.Enrolled = isEnrolled
		return manager.Storage.Update(pilot)
	}

	return manager.Save(&Pilot{FeatureFlagID: ff.ID, ExternalID: pilotID, Enrolled: isEnrolled})

}

//----------------------------------------------------------------------------------------------------------------------

func (manager *RolloutManager) ensureFeatureFlag(featureFlagName string) (*FeatureFlag, error) {

	ff, err := manager.Storage.FindFlagByName(featureFlagName)

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
			RandSeed: manager.RandSeedGenerator(),
			Strategy: RolloutStrategy{
				Percentage:       0,
				DecisionLogicAPI: nil,
			},
		},
	}
}
