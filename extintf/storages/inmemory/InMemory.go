package inmemory

import (
	"context"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources/storages/memorystorage"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
)

func New() *InMemory {
	return &InMemory{Memory: memorystorage.NewMemory()}
}

type InMemory struct{ *memorystorage.Memory }

func (s *InMemory) FindPilotEntriesByExtID(ctx context.Context, pilotExtID string) rollouts.PilotEntries {
	var pilots []*rollouts.Pilot

	for _, e := range s.TableFor(rollouts.Pilot{}) {
		p := e.(*rollouts.Pilot)

		if p.ExternalID == pilotExtID {
			pilots = append(pilots, p)
		}
	}

	return iterators.NewSlice(pilots)
}

func (s *InMemory) FindFlagsByName(ctx context.Context, names ...string) frameless.Iterator {
	var flags []*rollouts.FeatureFlag

	nameIndex := make(map[string]struct{})

	for _, name := range names {
		nameIndex[name] = struct{}{}
	}

	for _, e := range s.TableFor(rollouts.FeatureFlag{}) {
		flag := e.(*rollouts.FeatureFlag)

		if _, ok := nameIndex[flag.Name] ; ok {
			flags = append(flags, flag)
		}
	}

	return iterators.NewSlice(flags)
}

func (s *InMemory) FindPilotsByFeatureFlag(ctx context.Context, ff *rollouts.FeatureFlag) frameless.Iterator {
	table := s.TableFor(rollouts.Pilot{})

	var pilots []*rollouts.Pilot

	for _, v := range table {
		pilot := v.(*rollouts.Pilot)

		if pilot.FeatureFlagID == ff.ID {
			pilots = append(pilots, pilot)
		}
	}

	return iterators.NewSlice(pilots)
}

func (s *InMemory) FindFlagPilotByExternalPilotID(ctx context.Context, featureFlagID, externalPilotID string) (*rollouts.Pilot, error) {
	table := s.TableFor(rollouts.Pilot{})

	for _, v := range table {
		pilot := v.(*rollouts.Pilot)

		if pilot.FeatureFlagID == featureFlagID && pilot.ExternalID == externalPilotID {
			return pilot, nil
		}
	}

	return nil, nil
}

func (s *InMemory) FindFlagByName(ctx context.Context, name string) (*rollouts.FeatureFlag, error) {
	var ptr *rollouts.FeatureFlag
	table := s.TableFor(ptr)

	for _, v := range table {
		flag := v.(*rollouts.FeatureFlag)

		if flag.Name == name {
			ptr = flag
			break
		}
	}

	return ptr, nil
}

func (s *InMemory) FindTokenBySHA512Hex(ctx context.Context, t string) (*security.Token, error) {
	table := s.TableFor(security.Token{})

	for _, token := range table {
		token := token.(*security.Token)

		if token.SHA512 == t {
			return token, nil
		}
	}

	return nil, nil
}
