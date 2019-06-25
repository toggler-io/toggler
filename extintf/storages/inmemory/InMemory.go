package inmemory

import (
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources/storages/memorystorage"
)

func New() *InMemory {
	return &InMemory{Memory: memorystorage.NewMemory()}
}

type InMemory struct{ *memorystorage.Memory }

func (s *InMemory) FindPilotsByFeatureFlag(ff *rollouts.FeatureFlag) frameless.Iterator {
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

func (s *InMemory) FindFlagPilotByExternalPilotID(featureFlagID, externalPilotID string) (*rollouts.Pilot, error) {
	table := s.TableFor(rollouts.Pilot{})

	for _, v := range table {
		pilot := v.(*rollouts.Pilot)

		if pilot.FeatureFlagID == featureFlagID && pilot.ExternalID == externalPilotID {
			return pilot, nil
		}
	}

	return nil, nil
}

func (s *InMemory) FindFlagByName(name string) (*rollouts.FeatureFlag, error) {
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

func (s *InMemory) FindTokenByTokenString(tokenStr string) (*security.Token, error) {
	table := s.TableFor(security.Token{})

	for _, token := range table {
		token := token.(*security.Token)

		if token.Token == tokenStr {
			return token, nil
		}
	}

	return nil, nil
}
