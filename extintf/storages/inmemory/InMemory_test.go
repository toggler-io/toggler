package inmemory_test

import (
	"github.com/adamluzsi/toggler/extintf/storages/inmemory"
	testing2 "github.com/adamluzsi/toggler/testing"
	"github.com/adamluzsi/toggler/usecases/specs"
	"testing"
)

func TestInMemory(t *testing.T) {
	(&specs.StorageSpec{Subject: inmemory.New(), FixtureFactory: testing2.NewFixtureFactory()}).Test(t)
}
