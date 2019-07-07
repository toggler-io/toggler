package security

import (
	"context"

	"github.com/adamluzsi/frameless/resources/specs"
)

type Storage interface {
	specs.Save
	specs.FindByID
	specs.Truncate
	specs.DeleteByID
	specs.Update
	TokenFinder
}

type TokenFinder interface {
	FindTokenBySHA512Hex(ctx context.Context, sha512hex string) (*Token, error)
}
