package rollouts

// FeatureFlag is the basic entity with properties that feature flag holds
type FeatureFlag struct {
	// ID represent the fact that this object will be persistent in the Storage
	ID      string `ext:"ID"`
	Name    string
	Rollout Rollout
}

type Rollout struct {
	// RandSeedSalt allows you to configure the randomness for the percentage based pilot enrollment selection.
	// This value could have been neglected by using the flag name as random seed,
	// but that would reduce the flexibility for edge cases where you want
	// to use a similar pilot group as a successful flag rollout before.
	RandSeedSalt int64

	// Strategy expects to determines the behavior of the rollout workflow.
	// the actual behavior implementation is with the RolloutManager,
	// but the configuration data is located here
	Strategy RolloutStrategy
}

type RolloutStrategy struct {
	// Percentage allows you to define how many of your user base should be enrolled pseudo randomly.
	Percentage int
	// URL allow you to do rollout based on custom domain needs such as target groups,
	// which decision logic is available trough an API endpoint call
	URL string
}
