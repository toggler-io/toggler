package rollouts

// FeatureFlag is the basic entity with properties that feature flag holds
type FeatureFlag struct {
	ID      string `ext:"ID"`
	Name    string
	Rollout Rollout
}

type Rollout struct {
	GloballyEnabled bool
	Percentage      int
	URL             string
}
