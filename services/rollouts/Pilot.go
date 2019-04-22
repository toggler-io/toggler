package rollouts

type Pilot struct {
	ID string `ext:"ID"`
	// FeatureFlagID is the reference ID that can tell where this user record belongs to.
	FeatureFlagID string
	// ExternalID is the uniq id key that connect the caller services,
	// with this service and able to use A-B/Percentage or Pilot based testings.
	ExternalID string
	// Enrolled states that whether the Pilot for the given feature is enrolled, or blacklisted
	Enrolled bool
}
