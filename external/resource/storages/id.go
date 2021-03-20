package storages

import (
	"github.com/adamluzsi/frameless/resources"
	"github.com/google/uuid"
)

func EnsureID(ptr interface{}) error {
	if currentID, _ := resources.LookupID(ptr); currentID == `` {
		uuidV4, err := newV4UUID()
		if err != nil {
			return err
		}
		if err := resources.SetID(ptr, uuidV4); err != nil {
			return err
		}
	}
	return nil
}

func newV4UUID() (string, error) {
	uuidV4, err := uuid.NewRandom()
	if err != nil {
		return ``, err
	}
	return uuidV4.String(), nil
}

func isUUIDValid(id interface{}) bool {
	sid, ok := id.(string)
	if !ok {
		return false
	}
	_, err := uuid.Parse(sid)
	return err == nil
}
