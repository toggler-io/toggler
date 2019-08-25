package testing

import (
	"github.com/toggler-io/toggler/extintf/storages/inmemory"
)

func NewTestStorage() *TestStorage {
	return &TestStorage{InMemory: inmemory.New()}
}

type TestStorage struct{ *inmemory.InMemory }
