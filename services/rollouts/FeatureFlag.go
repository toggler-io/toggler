package rollouts

// FeatureFlag is the basic entity with properties that feature flag holds
type FeatureFlag struct {
	ID       string `ext:"ID"`
	Name     string
	IsGlobal bool
}

