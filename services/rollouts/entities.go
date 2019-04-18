package rollouts

// FeatureFlag is the basic entity with properties that feature flag holds
type FeatureFlag interface {
	Name() string
	IsGlobal() bool
	Pilots() []User
}

type User interface {
	ID() string
}