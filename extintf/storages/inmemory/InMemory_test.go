package inmemory_test

import (
	"github.com/adamluzsi/toggler/extintf/storages/inmemory"
	"github.com/adamluzsi/toggler/usecases/specs"
	"testing"
)

func TestInMemory(t *testing.T) {
	(&specs.StorageSpec{Subject: inmemory.New()}).Test(t)
}
