package release

// Flag is the basic entity with properties that feature flag holds
type Flag struct {
	ID   string `ext:"ID" json:"id,omitempty"`
	Name string `json:"name"`
}

func (f Flag) Validate() error {
	if f.Name == "" {
		return ErrNameIsEmpty
	}
	return nil
}
