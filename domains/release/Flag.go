package release

import (
	"net/url"
)

// Flag is the basic entity with properties that feature flag holds
type Flag struct {
	// ID represent the fact that this object will be persistent in the Subject
	ID      string      `ext:"ID" json:"id,omitempty"`
	Name    string      `json:"name"`
	Rollout FlagRollout `json:"rollout"`
}

type FlagRollout struct {
	// RandSeed allows you to configure the randomness for the percentage based pilot enrollment selection.
	// This value could have been neglected by using the flag name as random seed,
	// but that would reduce the flexibility for edge cases where you want
	// to use a similar pilot group as a successful flag rollout before.
	RandSeed int64 `json:"rand_seed_salt"`

	// Strategy expects to determines the behavior of the rollout workflow.
	// the actual behavior implementation is with the RolloutManager,
	// but the configuration data is located here
	Strategy FlagRolloutStrategy `json:"strategy"`
}

type FlagRolloutStrategy struct {
	// Percentage allows you to define how many of your user base should be enrolled pseudo randomly.
	Percentage int `json:"percentage"`
	// DecisionLogicAPI allow you to do rollout based on custom domain needs such as target groups,
	// which decision logic is available trough an API endpoint call
	DecisionLogicAPI *url.URL `json:"decision_logic_api"`
}

func (flag Flag) Verify() error {
	if flag.Name == "" {
		return ErrNameIsEmpty
	}

	if flag.Rollout.Strategy.Percentage < 0 || 100 < flag.Rollout.Strategy.Percentage {
		return ErrInvalidPercentage
	}

	if flag.Rollout.Strategy.DecisionLogicAPI != nil {
		_, err := url.ParseRequestURI(flag.Rollout.Strategy.DecisionLogicAPI.String())
		if err != nil {
			return ErrInvalidRequestURL
		}

		if flag.Rollout.Strategy.DecisionLogicAPI.Scheme == `` {
			return ErrInvalidRequestURL
		}

		if flag.Rollout.Strategy.DecisionLogicAPI.Hostname() == `` {
			return ErrInvalidRequestURL
		}
	}

	return nil
}

func (s FlagRolloutStrategy) IsGlobal() bool {
	return s.Percentage == 100
}
