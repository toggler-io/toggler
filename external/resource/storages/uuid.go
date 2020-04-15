package storages

import "github.com/google/uuid"

func newV4UUID() (string, error) {
	uuidV4, err := uuid.NewRandom()
	if err != nil {
		return ``, err
	}
	return uuidV4.String(), nil
}

func isUUIDValid(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
