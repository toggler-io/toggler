package contracts_test

import (
	c "github.com/adamluzsi/frameless/contracts"
	"github.com/toggler-io/toggler/domains/security/contracts"
)

var _ = []c.Interface{
	contracts.Storage{},
	contracts.TokenStorage{},
	contracts.TokenFinder{},
}
