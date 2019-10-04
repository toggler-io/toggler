package release

import (
	"context"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"net/http"
	"net/url"
	"time"
)

func NewFlagChecker(s Storage) *FlagChecker {
	return &FlagChecker{
		Storage:                s,
		IDPercentageCalculator: PseudoRandPercentageGenerator{}.FNV1a64,
		HTTPClient:             http.Client{Timeout: 3 * time.Second},
	}
}

// FlagChecker is an interactor that implements query like (read only) behaviors with feature flags.
type FlagChecker struct {
	Storage                Storage
	IDPercentageCalculator func(string, int64) (int, error)
	HTTPClient             http.Client
}

// IsFeatureEnabledFor grant you the ability to
// check whether a pilot is enrolled or not for the feature flag in subject.
func (checker *FlagChecker) IsFeatureEnabledFor(featureFlagName string, externalPilotID string) (bool, error) {

	ff, err := checker.Storage.FindReleaseFlagByName(context.TODO(), featureFlagName)

	if err != nil {
		return false, err
	}

	if ff == nil {
		return false, nil
	}

	pilot, err := checker.Storage.FindReleaseFlagPilotByPilotExternalID(context.TODO(), ff.ID, externalPilotID)

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

	return checker.isEnrolled(ff, externalPilotID)

}

func (checker *FlagChecker) IsFeatureGloballyEnabled(featureFlagName string) (bool, error) {
	ff, err := checker.Storage.FindReleaseFlagByName(context.TODO(), featureFlagName)

	if err != nil {
		return false, err
	}

	if ff == nil {
		return false, nil
	}

	return ff.Rollout.Strategy.Percentage == 100, nil
}

func (checker *FlagChecker) makeCustomDomainAPIDecisionCheck(featureFlagName string, externalPilotID string, apiURL *url.URL) (bool, error) {
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

func (checker *FlagChecker) GetPilotFlagStates(ctx context.Context, externalPilotID string, featureFlagNames ...string) (map[string]bool, error) {
	states := make(map[string]bool)

	for _, flagName := range featureFlagNames {
		states[flagName] = false
	}

	flagIndexByID := make(map[string]*Flag)

	flags := checker.Storage.FindReleaseFlagsByName(ctx, featureFlagNames...)

	for flags.Next() {
		var ff Flag
		if err := flags.Decode(&ff); err != nil {
			return nil, err
		}
		flagIndexByID[ff.ID] = &ff

		enrollment, err := checker.isEnrolled(&ff, externalPilotID)

		if err != nil {
			return nil, err
		}

		states[ff.Name] = enrollment
	}

	pilotEntries := checker.Storage.FindPilotEntriesByExtID(ctx, externalPilotID)
	pilotEntriesThatAreRelevant := iterators.Filter(pilotEntries, func(p frameless.Entity) bool {
		_, ok := flagIndexByID[p.(Pilot).FlagID]
		return ok
	})

	for pilotEntriesThatAreRelevant.Next() {
		var p Pilot
		if err := pilotEntriesThatAreRelevant.Decode(&p); err != nil {
			return nil, err
		}
		states[flagIndexByID[p.FlagID].Name] = p.Enrolled
	}

	return states, nil
}

func (checker *FlagChecker) isEnrolled(ff *Flag, externalPilotID string) (bool, error) {
	diceRollResultPercentage, err := checker.IDPercentageCalculator(externalPilotID, ff.Rollout.RandSeed)
	if err != nil {
		return false, err
	}
	return diceRollResultPercentage <= ff.Rollout.Strategy.Percentage, nil
}
