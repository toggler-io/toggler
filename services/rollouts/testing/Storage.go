package testing

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources/storages/memorystorage"
)

func NewStorage() *Storage {
	return &Storage{Memory: memorystorage.NewMemory()}
}

type Storage struct{ *memorystorage.Memory }

func (storage *Storage) FindPilotsByFeatureFlag(ff *rollouts.FeatureFlag) frameless.Iterator {
	table := storage.TableFor(rollouts.Pilot{})

	var pilots []*rollouts.Pilot

	for _, v := range table {
		pilot := v.(*rollouts.Pilot)

		if pilot.FeatureFlagID == ff.ID {
			pilots = append(pilots, pilot)
		}
	}

	return iterators.NewSlice(pilots)
}

//TODO: fix name here
func (storage *Storage) FindFlagPilotByExternalPilotID(featureFlagID, externalPilotID string) (*rollouts.Pilot, error) {
	table := storage.TableFor(rollouts.Pilot{})

	for _, v := range table {
		pilot := v.(*rollouts.Pilot)

		if pilot.FeatureFlagID == featureFlagID && pilot.ExternalID == externalPilotID {
			return pilot, nil
		}
	}

	return nil, nil
}

func (storage *Storage) FindByFlagName(name string) (*rollouts.FeatureFlag, error) {
	var ptr *rollouts.FeatureFlag
	table := storage.TableFor(ptr)

	for _, v := range table {
		flag := v.(*rollouts.FeatureFlag)

		if flag.Name == name {
			ptr = flag
			break
		}
	}

	return ptr, nil
}
