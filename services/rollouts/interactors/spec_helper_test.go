package interactors_test

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts/testing"
)

func NewTestStorage() *testing.Storage {
	return testing.NewStorage()
}
