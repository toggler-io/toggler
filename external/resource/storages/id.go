package storages

import (
	"context"
	"github.com/adamluzsi/frameless/extid"
	"github.com/google/uuid"
)

func EnsureID(ptr interface{}) error {
	if currentID, _ := extid.Lookup(ptr); currentID == `` {
		uuidV4, err := newV4UUID()
		if err != nil {
			return err
		}
		if err := extid.Set(ptr, uuidV4); err != nil {
			return err
		}
	}
	return nil
}

func newIDFn(ctx context.Context) (interface{}, error) {
	return newV4UUID()
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
