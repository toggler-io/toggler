package rollouts

import (
	"net/url"
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

func (manager *RolloutManager) SetFeatureFlag(flag *FeatureFlag) error {
	if flag == nil {
		return ErrMissingFlag
	}

	if flag.Name == "" {
		return ErrInvalidFeatureName
	}

	if flag.Rollout.Strategy.DecisionLogicAPI != nil {
		_, err := url.ParseRequestURI(flag.Rollout.Strategy.DecisionLogicAPI.String())
		if err != nil {
			return ErrInvalidURL
		}
	}

	if flag.Rollout.Strategy.Percentage < 0 || 100 < flag.Rollout.Strategy.Percentage {
		return ErrInvalidPercentage
	}

	if flag.ID == `` {
		ff, err := manager.Storage.FindFlagByName(flag.Name)
		if err != nil {
			return err
		}

		if ff != nil {
			flag.ID = ff.ID

			flag.Rollout.RandSeedSalt = ff.Rollout.RandSeedSalt
		}
	}

	if flag.Rollout.RandSeedSalt == 0 {
		flag.Rollout.RandSeedSalt = manager.RandSeedGenerator()
	}

	var persister func(interface{}) error

	if flag.ID == ``{
		persister = manager.Storage.Save
	} else {
		persister = manager.Storage.Update
	}

	return persister(flag)
}

func (manager *RolloutManager) ListFeatureFlags() ([]*FeatureFlag, error) {
	iter := manager.Storage.FindAll(FeatureFlag{})
	ffs := []*FeatureFlag{} // empty slice required for null object pattern enforcement
	err := iterators.CollectAll(iter, &ffs)
	return ffs, err
}

func (manager *RolloutManager) SetPilotEnrollmentForFeature(featureFlagName string, pilotExtID string, isEnrolled bool) error {

	ff, err := manager.ensureFeatureFlag(featureFlagName)

	if err != nil {
		return err
	}

	pilot, err := manager.Storage.FindFlagPilotByExternalPilotID(ff.ID, pilotExtID)

	if err != nil {
		return err
	}

	if pilot != nil {
		pilot.Enrolled = isEnrolled
		return manager.Storage.Update(pilot)
	}

	return manager.Save(&Pilot{FeatureFlagID: ff.ID, ExternalID: pilotExtID, Enrolled: isEnrolled})

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
			RandSeedSalt: manager.RandSeedGenerator(),
			Strategy: RolloutStrategy{
				Percentage:       0,
				DecisionLogicAPI: nil,
			},
		},
	}
}
