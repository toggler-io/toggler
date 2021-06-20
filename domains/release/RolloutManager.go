package release

import (
	"context"
	"fmt"

	"github.com/adamluzsi/frameless/iterators"

	"github.com/toggler-io/toggler/domains/deployment"
)

func NewRolloutManager(s Storage) *RolloutManager {
	return &RolloutManager{Storage: s}
}

// RolloutManager provides you with feature flag configurability.
// The manager use storage in a write heavy behavior.
//
// SRP: release manager
type RolloutManager struct{ Storage Storage }

// GetAllReleaseFlagStatesOfThePilot check the flag states for every requested release flag.
// If a flag doesn't exist, then it will provide a turned off state to it.
// This helps with development where the flag name already agreed and used in the clients
// but the entity not yet been created in the system.
// Also this help if a flag is not cleaned up from the clients, the worst thing will be a disabled feature,
// instead of a breaking client.
// This also makes it harder to figure out `private` release flags
func (manager *RolloutManager) GetAllReleaseFlagStatesOfThePilot(ctx context.Context, pilotExternalID string, env deployment.Environment, flagNames ...string) (map[string]bool, error) {
	states := make(map[string]bool)

	for _, flagName := range flagNames {
		states[flagName] = false
	}

	flagIndexByID := make(map[string]*Flag)

	var pilotsIndex = make(map[string]*ManualPilot)
	pilotsByExternalID := manager.Storage.ReleasePilot(ctx).FindReleasePilotsByExternalID(ctx, pilotExternalID)
	pilotsByExternalIDFilteredByEnv := iterators.Filter(pilotsByExternalID, func(p ManualPilot) bool {
		return p.DeploymentEnvironmentID == env.ID
	})
	if err := iterators.ForEach(pilotsByExternalIDFilteredByEnv, func(p ManualPilot) error {
		pilotsIndex[p.FlagID] = &p
		return nil
	}); err != nil {
		return nil, err
	}

	var flags []Flag
	if err := iterators.Collect(manager.Storage.ReleaseFlag(ctx).FindReleaseFlagsByName(ctx, flagNames...), &flags); err != nil {
		return nil, err
	}

	for _, f := range flags {
		flagIndexByID[f.ID] = &f

		if p, ok := pilotsIndex[f.ID]; ok {
			states[f.Name] = p.IsParticipating
			continue
		}

		enrollment, err := manager.checkEnrollment(ctx, env, f, pilotExternalID, pilotsIndex)
		if err != nil {
			return nil, err
		}

		states[f.Name] = enrollment
	}

	return states, nil
}

func (manager *RolloutManager) checkEnrollment(ctx context.Context, env deployment.Environment, flag Flag, pilotExternalID string, manualPilotEnrollmentIndex map[string]*ManualPilot) (bool, error) {
	if p, ok := manualPilotEnrollmentIndex[flag.ID]; ok {
		return p.IsParticipating, nil
	}

	var rollout Rollout
	found, err := manager.Storage.ReleaseRollout(ctx).FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx, flag, env, &rollout)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}

	return rollout.Plan.IsParticipating(ctx, pilotExternalID)
}

func (manager *RolloutManager) CreateFeatureFlag(ctx context.Context, flag *Flag) error {
	if flag == nil {
		return ErrMissingFlag
	}

	if err := flag.Validate(); err != nil {
		return err
	}

	if flag.ID != `` {
		return ErrInvalidAction
	}
	ff, err := manager.Storage.ReleaseFlag(ctx).FindReleaseFlagByName(ctx, flag.Name)

	if err != nil {
		return err
	}

	if ff != nil {
		//TODO: this should be handled in transaction!
		// as mvp solution, it is acceptable for now,
		// but spec must be moved to the storage specs as `name is uniq across entries`
		// or transaction through context with serialization level must be used for this action
		return ErrFlagAlreadyExist
	}

	return manager.Storage.ReleaseFlag(ctx).Create(ctx, flag)
}

func (manager *RolloutManager) UpdateFeatureFlag(ctx context.Context, flag *Flag) error {
	if flag == nil {
		return ErrMissingFlag
	}

	if err := flag.Validate(); err != nil {
		return err
	}

	return manager.Storage.ReleaseFlag(ctx).Update(ctx, flag)
}

// TODO convert this into a stream
func (manager *RolloutManager) ListFeatureFlags(ctx context.Context) ([]Flag, error) {
	iter := manager.Storage.ReleaseFlag(ctx).FindAll(ctx)
	ffs := make([]Flag, 0) // empty slice required for null object pattern enforcement
	err := iterators.Collect(iter, &ffs)
	return ffs, err
}

func (manager *RolloutManager) UnsetPilotEnrollmentForFeature(ctx context.Context, flagID string, envID string, pilotExternalID string) error {

	var ff Flag

	found, err := manager.Storage.ReleaseFlag(ctx).FindByID(ctx, &ff, flagID)

	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf(`ErrNotFound`)
	}

	pilot, err := manager.Storage.ReleasePilot(ctx).FindReleaseManualPilotByExternalID(ctx, ff.ID, envID, pilotExternalID)

	if err != nil {
		return err
	}

	if pilot == nil {
		return nil
	}

	return manager.Storage.ReleasePilot(ctx).DeleteByID(ctx, pilot.ID)

}

func (manager *RolloutManager) SetPilotEnrollmentForFeature(ctx context.Context, flagID string, envID string, externalPilotID string, isParticipating bool) error {

	var ff Flag

	found, err := manager.Storage.ReleaseFlag(ctx).FindByID(ctx, &ff, flagID)

	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf(`ErrNotFound`)
	}

	pilot, err := manager.Storage.ReleasePilot(ctx).FindReleaseManualPilotByExternalID(ctx, ff.ID, envID, externalPilotID)

	if err != nil {
		return err
	}

	if pilot != nil {
		pilot.IsParticipating = isParticipating
		return manager.Storage.ReleasePilot(ctx).Update(ctx, pilot)
	}

	return manager.Storage.ReleasePilot(ctx).Create(ctx, &ManualPilot{
		FlagID:                  ff.ID,
		DeploymentEnvironmentID: envID,
		ExternalID:              externalPilotID,
		IsParticipating:         isParticipating,
	})

}

//TODO: make operation atomic between flags and pilots
// TODO delete ip addr allows as well
// TODO: rename
func (manager *RolloutManager) DeleteFeatureFlag(ctx context.Context, id string) (returnErr error) {
	ctx, err := manager.Storage.BeginTx(ctx)
	if err != nil {
		return err
	}
	var noRollback bool
	defer func() {
		if returnErr != nil {
			if noRollback {
				return
			}

			_ = manager.Storage.RollbackTx(ctx)
			return
		}

		returnErr = manager.Storage.CommitTx(ctx)
	}()

	if id == `` {
		noRollback = true
		return fmt.Errorf(`entity id is empty`)
	}

	var ff Flag
	found, err := manager.Storage.ReleaseFlag(ctx).FindByID(ctx, &ff, id)

	if err != nil {
		return err
	}
	if !found {
		noRollback = true
		return fmt.Errorf(`ErrNotFound`)
	}

	var pilots []ManualPilot
	if err := iterators.Collect(manager.Storage.ReleasePilot(ctx).FindReleasePilotsByReleaseFlag(ctx, ff), &pilots); err != nil {
		return err
	}

	for _, pilot := range pilots {
		if err := manager.Storage.ReleasePilot(ctx).DeleteByID(ctx, pilot.ID); err != nil {
			return err
		}
	}

	return manager.Storage.ReleaseFlag(ctx).DeleteByID(ctx, id)
}
