package deployment

type Environment struct {
	ID   string `ext:"ID" json:"id"`
	Name string `json:"name"`
}

func (env Environment) Validate() error {
	if env.Name == "" {
		return ErrEnvironmentNameIsEmpty
	}
	return nil
}
