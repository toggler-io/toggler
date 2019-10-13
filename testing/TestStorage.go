package testing

import (
	"github.com/toggler-io/toggler/extintf/storages"
)

func NewTestStorage() *TestStorage {
	return &TestStorage{InMemory: storages.NewInMemory()}
}

type TestStorage struct{ *storages.InMemory }
