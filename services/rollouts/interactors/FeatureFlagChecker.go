package interactors

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"net/http"
	"time"
)

func NewFeatureFlagChecker(s rollouts.Storage) *FeatureFlagChecker {
	return &FeatureFlagChecker{
		Storage:                s,
		IDPercentageCalculator: GeneratePseudoRandPercentageWithFNV1a64,
		HTTPClient:             http.Client{Timeout: 3 * time.Second},
	}
}

type FeatureFlagChecker struct {
	Storage                rollouts.Storage
	IDPercentageCalculator func(string, int64) (int, error)
	HTTPClient             http.Client
}

func (checker *FeatureFlagChecker) IsFeatureEnabledFor(featureFlagName string, externalPilotID string) (bool, error) {

	ff, err := checker.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return false, err
	}

	if ff == nil {
		return false, nil
	}

	pilot, err := checker.Storage.FindFlagPilotByExternalPilotID(ff.ID, externalPilotID)

	if err != nil {
		return false, err
	}

	if pilot != nil {
		return pilot.Enrolled, nil
	}

	if ff.Rollout.Strategy.URL != "" {
		return checker.makeCustomDomainAPIDecisionCheck(featureFlagName, externalPilotID, ff.Rollout.Strategy.URL)
	}

	if ff.Rollout.Strategy.Percentage == 0 {
		return false, nil
	}

	diceRollResultPercentage, err := checker.IDPercentageCalculator(externalPilotID, ff.Rollout.RandSeedSalt)

	if err != nil {
		return false, err
	}

	return diceRollResultPercentage <= ff.Rollout.Strategy.Percentage, nil

}

func (checker *FeatureFlagChecker) IsFeatureGloballyEnabled(featureFlagName string) (bool, error) {
	ff, err := checker.Storage.FindByFlagName(featureFlagName)

	if err != nil {
		return false, err
	}

	if ff == nil {
		return false, nil
	}

	return ff.Rollout.Strategy.Percentage == 100, nil
}

func (checker *FeatureFlagChecker) makeCustomDomainAPIDecisionCheck(featureFlagName string, externalPilotID string, url string) (bool, error) {
	req, err := http.NewRequest(`GET`, url, nil)
	if err != nil {
		return false, err
	}

	query := req.URL.Query()
	query.Set(`feature-flag-name`, featureFlagName)
	query.Set(`pilot-id`, externalPilotID)
	req.URL.RawQuery = query.Encode()

	resp, err := checker.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}

	code := resp.StatusCode
	return 200 <= code && code < 300, nil
}
