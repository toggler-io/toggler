package deployment

type Environment struct {
	ID   string `ext:"ID"`
	Name string
}

func (env Environment) Validate() error {
	if env.Name == "" {
		return ErrEnvironmentNameIsEmpty
	}
	return nil
}
