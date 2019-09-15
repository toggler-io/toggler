package release

// Pilot is a data entity that represent relation between an external system's user and a feature flag.
// The Pilot terminology itself means that the user is in charge to try out a given feature,
// even if the user itself is not aware of this role.
type Pilot struct {
	// ID represent the fact that this object will be persistent in the Subject
	ID string `ext:"ID"`
	// FlagID is the reference ID that can tell where this user record belongs to.
	FlagID string
	// ExternalID is the uniq id key that connect the caller services,
	// with this service and able to use A-B/Percentage or Pilot based testings.
	ExternalID string
	// Enrolled states that whether the Pilot for the given feature is enrolled, or blacklisted
	Enrolled bool
}
