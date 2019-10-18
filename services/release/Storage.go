package release

import (
	"context"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/resources"
)

type Storage interface {
	resources.Saver
	resources.Finder
	resources.Updater
	resources.Deleter
	resources.Truncater
	FlagFinder
	PilotFinder
	AllowFinder
}

type AllowEntries = frameless.Iterator
type PilotEntries = frameless.Iterator
type FlagEntries = frameless.Iterator

type FlagFinder interface {
	FindReleaseFlagByName(ctx context.Context, name string) (*Flag, error)
	FindReleaseFlagsByName(ctx context.Context, names ...string) FlagEntries
}

type PilotFinder interface {
	FindReleaseFlagPilotByPilotExternalID(ctx context.Context, flagID, pilotExtID string) (*Pilot, error)
	FindPilotsByFeatureFlag(ctx context.Context, ff *Flag) PilotEntries
	FindPilotEntriesByExtID(ctx context.Context, pilotExtID string) PilotEntries
}

type AllowFinder interface {
	FindReleaseAllowsByReleaseFlags(ctx context.Context, flags ...*Flag) AllowEntries
}
