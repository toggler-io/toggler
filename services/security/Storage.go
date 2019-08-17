package security

import (
	"context"

	"github.com/adamluzsi/frameless/resources"
)

type Storage interface {
	resources.Save
	resources.FindByID
	resources.Truncate
	resources.DeleteByID
	resources.Update
	TokenFinder
}

type TokenFinder interface {
	FindTokenBySHA512Hex(ctx context.Context, sha512hex string) (*Token, error)
}
