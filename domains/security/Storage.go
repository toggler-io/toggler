package security

import (
	"context"

	"github.com/adamluzsi/frameless"
)

type Storage interface {
	frameless.OnePhaseCommitProtocol
	SecurityToken(context.Context) TokenStorage
}

type TokenStorage interface {
	frameless.Creator
	frameless.Finder
	frameless.Updater
	frameless.Deleter
	frameless.CreatorPublisher
	frameless.UpdaterPublisher
	frameless.DeleterPublisher
	FindTokenBySHA512Hex(ctx context.Context, sha512hex string) (*Token, error)
}
