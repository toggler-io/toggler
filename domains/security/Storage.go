package security

import (
	"context"

	"github.com/adamluzsi/frameless/resources"
)

type Storage interface {
	resources.Creator
	resources.Finder
	resources.Updater
	resources.Deleter
	TokenFinder

	resources.CreatorPublisher
	resources.UpdaterPublisher
	resources.DeleterPublisher
}

type TokenFinder interface {
	FindTokenBySHA512Hex(ctx context.Context, sha512hex string) (*Token, error)
}
