package rollouts

import (
	"github.com/adamluzsi/frameless"
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

const ErrInvalidPercentage frameless.Error = `percentage value not acceptable`

func (manager *RolloutManager) UpdateFeatureFlagRolloutPercentage(featureFlagName string, rolloutPercentage int) error {

	if rolloutPercentage < 0 || 100 < rolloutPercentage {
		return ErrInvalidPercentage
	}

	ff, err := manager.Storage.FindFlagByName(featureFlagName)

	if err != nil {
		return err
	}

	if ff == nil {
		ff = manager.newDefaultFeatureFlag(featureFlagName)
		ff.Rollout.Strategy.Percentage = rolloutPercentage
		return manager.Storage.Save(ff)
	}

	ff.Rollout.Strategy.Percentage = rolloutPercentage
	return manager.Storage.Update(ff)

}

func (manager *RolloutManager) ListFeatureFlags() ([]*FeatureFlag, error) {
	iter := manager.Storage.FindAll(FeatureFlag{})
	ffs := []*FeatureFlag{} // empty slice required for null object pattern enforcement
	err := iterators.CollectAll(iter, &ffs)
	return ffs, err
}

const ErrInvalidURL frameless.Error = `url value not acceptable`

func (manager *RolloutManager) SetFeatureFlagRolloutStrategyToUseDecisionLogicAPI(featureFlagName string, decisionAPIURL *url.URL) error {

	if decisionAPIURL == nil {
		return ErrInvalidURL
	}

	ff, err := manager.Storage.FindFlagByName(featureFlagName)

	if err != nil {
		return err
	}

	if ff == nil {
		ff = manager.newDefaultFeatureFlag(featureFlagName)
		ff.Rollout.Strategy.DecisionLogicAPI = decisionAPIURL
		return manager.Storage.Save(ff)
	}

	ff.Rollout.Strategy.DecisionLogicAPI = decisionAPIURL
	return manager.Storage.Update(ff)

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
