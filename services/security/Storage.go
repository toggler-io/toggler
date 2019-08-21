package security

import (
	"context"

	"github.com/adamluzsi/frameless/resources"
)

type Storage interface {
	resources.Saver
	resources.Finder
	resources.Updater
	resources.Deleter
	resources.Truncater
	TokenFinder
}

type TokenFinder interface {
	FindTokenBySHA512Hex(ctx context.Context, sha512hex string) (*Token, error)
}
