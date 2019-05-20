package rollouts

import (
	"net/http"
	"net/url"
	"time"
)

func NewFeatureFlagChecker(s Storage) *FeatureFlagChecker {
	return &FeatureFlagChecker{
		Storage:                s,
		IDPercentageCalculator: GeneratePseudoRandPercentageWithFNV1a64,
		HTTPClient:             http.Client{Timeout: 3 * time.Second},
	}
}

// FeatureFlagChecker is an interactor that implements query like (read only) behaviors with feature flags.
type FeatureFlagChecker struct {
	Storage                Storage
	IDPercentageCalculator func(string, int64) (int, error)
	HTTPClient             http.Client
}

// IsFeatureEnabledFor grant you the ability to
// check whether a pilot is enrolled or not for the feature flag in subject.
func (checker *FeatureFlagChecker) IsFeatureEnabledFor(featureFlagName string, externalPilotID string) (bool, error) {

	ff, err := checker.Storage.FindFlagByName(featureFlagName)

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

	if ff.Rollout.Strategy.DecisionLogicAPI != nil {
		return checker.makeCustomDomainAPIDecisionCheck(featureFlagName, externalPilotID, ff.Rollout.Strategy.DecisionLogicAPI)
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
	ff, err := checker.Storage.FindFlagByName(featureFlagName)

	if err != nil {
		return false, err
	}

	if ff == nil {
		return false, nil
	}

	return ff.Rollout.Strategy.Percentage == 100, nil
}

func (checker *FeatureFlagChecker) makeCustomDomainAPIDecisionCheck(featureFlagName string, externalPilotID string, apiURL *url.URL) (bool, error) {
	req, err := http.NewRequest(`GET`, apiURL.String(), nil)
	if err != nil {
		return false, err
	}

	query := req.URL.Query()
	query.Set(`feature`, featureFlagName)
	query.Set(`id`, externalPilotID)
	req.URL.RawQuery = query.Encode()

	resp, err := checker.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}

	code := resp.StatusCode
	return 200 <= code && code < 300, nil
}
