package storages

import (
	"context"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources/storages/memorystorage"

	"github.com/toggler-io/toggler/services/release"
	"github.com/toggler-io/toggler/services/security"
)

func NewInMemory() *InMemory {
	return &InMemory{Memory: memorystorage.NewMemory()}
}

type InMemory struct{ *memorystorage.Memory }

func (s *InMemory) FindReleaseAllowsByReleaseFlags(ctx context.Context, flags ...*release.Flag) release.AllowEntries {
	var allows []release.IPAllow

	flagIndex := make(map[string]struct{})
	for _, flag := range flags {
		flagIndex[flag.ID] = struct{}{}
	}

	for _, e := range s.TableFor(release.IPAllow{}) {
		allow := e.(*release.IPAllow)
		if _, ok := flagIndex[allow.FlagID]; ok {
			allows = append(allows, *allow)
		}
	}

	return iterators.NewSlice(allows)
}

func (s *InMemory) FindPilotEntriesByExtID(ctx context.Context, pilotExtID string) release.PilotEntries {
	var pilots []release.Pilot

	for _, e := range s.TableFor(release.Pilot{}) {
		p := e.(*release.Pilot)

		if p.ExternalID == pilotExtID {
			pilots = append(pilots, *p)
		}
	}

	return iterators.NewSlice(pilots)
}

func (s *InMemory) FindReleaseFlagsByName(ctx context.Context, names ...string) frameless.Iterator {
	var flags []release.Flag

	nameIndex := make(map[string]struct{})

	for _, name := range names {
		nameIndex[name] = struct{}{}
	}

	for _, e := range s.TableFor(release.Flag{}) {
		flag := e.(*release.Flag)

		if _, ok := nameIndex[flag.Name] ; ok {
			flags = append(flags, *flag)
		}
	}

	return iterators.NewSlice(flags)
}

func (s *InMemory) FindPilotsByFeatureFlag(ctx context.Context, ff *release.Flag) frameless.Iterator {
	table := s.TableFor(release.Pilot{})

	var pilots []release.Pilot

	for _, v := range table {
		pilot := v.(*release.Pilot)

		if pilot.FlagID == ff.ID {
			pilots = append(pilots, *pilot)
		}
	}

	return iterators.NewSlice(pilots)
}

func (s *InMemory) FindReleaseFlagPilotByPilotExternalID(ctx context.Context, featureFlagID, externalPilotID string) (*release.Pilot, error) {
	table := s.TableFor(release.Pilot{})

	for _, v := range table {
		pilot := v.(*release.Pilot)

		if pilot.FlagID == featureFlagID && pilot.ExternalID == externalPilotID {
			p := *pilot
			return &p, nil
		}
	}

	return nil, nil
}

func (s *InMemory) FindReleaseFlagByName(ctx context.Context, name string) (*release.Flag, error) {
	for _, v := range s.TableFor(release.Flag{}) {
		flagRecord := v.(*release.Flag)

		if flagRecord.Name == name {
			f := *flagRecord
			return &f, nil
		}
	}
	return nil, nil
}

func (s *InMemory) FindTokenBySHA512Hex(ctx context.Context, t string) (*security.Token, error) {
	table := s.TableFor(security.Token{})

	for _, token := range table {
		token := token.(*security.Token)

		if token.SHA512 == t {
			t := *token
			return &t, nil
		}
	}

	return nil, nil
}
