package release

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
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

func (checker *FlagChecker) GetPilotFlagStates(ctx context.Context, pilotExtID string, flagNames ...string) (map[string]bool, error) {
	states := make(map[string]bool)

	for _, flagName := range flagNames {
		states[flagName] = false
	}

	flagIndexByID := make(map[string]*Flag)

	flags := checker.Storage.FindReleaseFlagsByName(ctx, flagNames...)

	pilotsIter := checker.Storage.FindPilotEntriesByExtID(ctx, pilotExtID)
	defer pilotsIter.Close()

	var pilotsIndex = make(map[string]*Pilot)
	for pilotsIter.Next() {
		var p Pilot
		if err := pilotsIter.Decode(&p); err != nil {
			return nil, err
		}
		pilotsIndex[p.FlagID] = &p
	}
	if err := pilotsIter.Err(); err != nil {
		return nil, err
	}

	for flags.Next() {
		var ff Flag
		if err := flags.Decode(&ff); err != nil {
			return nil, err
		}
		flagIndexByID[ff.ID] = &ff

		enrollment, err := checker.checkEnrollment(&ff, pilotExtID, pilotsIndex)

		if err != nil {
			return nil, err
		}

		states[ff.Name] = enrollment
	}

	pilotEntries := checker.Storage.FindPilotEntriesByExtID(ctx, pilotExtID)
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

func (checker *FlagChecker) checkEnrollment(flag *Flag, pilotExtID string, pilotIndex map[string]*Pilot) (bool, error) {

	if p, ok := pilotIndex[flag.ID]; ok {
		return p.Enrolled, nil
	}

	if flag.Rollout.Strategy.DecisionLogicAPI != nil {
		return checker.makeCustomDomainAPIDecisionCheck(flag.Name, pilotExtID, flag.Rollout.Strategy.DecisionLogicAPI)
	}

	if flag.Rollout.Strategy.Percentage == 0 {
		return false, nil
	}

	diceRollResultPercentage, err := checker.IDPercentageCalculator(pilotExtID, flag.Rollout.RandSeed)
	if err != nil {
		return false, err
	}

	return diceRollResultPercentage <= flag.Rollout.Strategy.Percentage, nil

}