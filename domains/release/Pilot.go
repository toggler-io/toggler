package release

// Pilot is a data entity that represent relation between an external system's user and a feature flag.
// The Pilot terminology itself means that the user is in charge to try out a given feature,
// even if the user itself is not aware of this role.
type Pilot struct {
	// ID represent the fact that this object will be persistent in the Subject
	ID string `ext:"ID"`
	// FlagID is the reference ID that can tell where this user record belongs to.
	FlagID string
	// EnvironmentID is the ID of the environment where the pilot should be enrolled
	EnvironmentID string
	// PublicID is the unique id that connects the entry to the caller services,
	// with this service and able to use A-B/Percentage or Pilot based testings.
	PublicID string
	// IsParticipating states that whether the pilot for the given flag in a given environment is enrolled, or blacklisted.
	IsParticipating bool
}
