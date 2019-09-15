package release

import (
	"context"
	"time"

	"github.com/adamluzsi/frameless"

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

func (manager *RolloutManager) CreateFeatureFlag(ctx context.Context, flag *Flag) error {
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

	ff, err := manager.Storage.FindReleaseFlagByName(ctx, flag.Name)

	if err != nil {
		return err
	}

	if ff != nil {
		//TODO: this should be handled in transaction!
		// as mvp solution, it is acceptable for now,
		// but spec must be moved to the storage specs as `name is uniq across entries`
		return ErrFlagAlreadyExist
	}

	return manager.Storage.Save(ctx, flag)
}

func (manager *RolloutManager) UpdateFeatureFlag(ctx context.Context, flag *Flag) error {
	if flag == nil {
		return ErrMissingFlag
	}

	if err := flag.Verify(); err != nil {
		return err
	}

	if flag.ID == `` {
		ff, err := manager.Storage.FindReleaseFlagByName(ctx, flag.Name)
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

	return manager.Storage.Update(ctx, flag)
}

func (manager *RolloutManager) ListFeatureFlags(ctx context.Context) ([]*Flag, error) {
	iter := manager.Storage.FindAll(ctx, Flag{})
	ffs := []*Flag{} // empty slice required for null object pattern enforcement
	err := iterators.CollectAll(iter, &ffs)
	return ffs, err
}

func (manager *RolloutManager) UnsetPilotEnrollmentForFeature(ctx context.Context, flagID, externalPilotID string) error {

	var ff Flag

	found, err := manager.Storage.FindByID(ctx, &ff, flagID)

	if err != nil {
		return err
	}

	if !found {
		return frameless.ErrNotFound
	}

	pilot, err := manager.Storage.FindReleaseFlagPilotByPilotExternalID(ctx, ff.ID, externalPilotID)

	if err != nil {
		return err
	}

	if pilot == nil {
		return nil
	}

	return manager.Storage.DeleteByID(ctx, pilot, pilot.ID)
	
}

func (manager *RolloutManager) SetPilotEnrollmentForFeature(ctx context.Context, flagID, externalPilotID string, isEnrolled bool) error {

	var ff Flag

	found, err := manager.Storage.FindByID(ctx, &ff, flagID)

	if err != nil {
		return err
	}

	if !found {
		return frameless.ErrNotFound
	}

	pilot, err := manager.Storage.FindReleaseFlagPilotByPilotExternalID(ctx, ff.ID, externalPilotID)

	if err != nil {
		return err
	}

	if pilot != nil {
		pilot.Enrolled = isEnrolled
		return manager.Storage.Update(ctx, pilot)
	}

	return manager.Save(ctx, &Pilot{FlagID: ff.ID, ExternalID: externalPilotID, Enrolled: isEnrolled})

}

//TODO: replace with Tx
func (manager *RolloutManager) DeleteFeatureFlag(ctx context.Context, id string) error {
	if id == `` {
		return frameless.ErrIDRequired
	}

	var ff Flag
	found, err := manager.Storage.FindByID(ctx, &ff, id)

	if err != nil {
		return err
	}

	if !found {
		return frameless.ErrNotFound
	}

	pilots := manager.Storage.FindPilotsByFeatureFlag(ctx, &ff)
	defer pilots.Close()

	for pilots.Next() {
		var pilot Pilot

		if err := pilots.Decode(&pilot); err != nil {
			return err
		}

		if err := manager.Storage.DeleteByID(ctx, pilot, pilot.ID); err != nil {
			return err
		}
	}

	if err := pilots.Err(); err != nil {
		return err
	}

	return manager.Storage.DeleteByID(ctx, Flag{}, id)
}

//----------------------------------------------------------------------------------------------------------------------

func (manager *RolloutManager) newDefaultFeatureFlag(featureFlagName string) *Flag {
	return &Flag{
		Name: featureFlagName,
		Rollout: FlagRollout{
			RandSeed: manager.RandSeedGenerator(),
			Strategy: FlagRolloutStrategy{
				Percentage:       0,
				DecisionLogicAPI: nil,
			},
		},
	}
}
