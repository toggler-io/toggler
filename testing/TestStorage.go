package testing

import (
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources/storages/memorystorage"
)

func NewTestStorage() *TestStorage {
	return &TestStorage{Memory: memorystorage.NewMemory()}
}

type TestStorage struct{ *memorystorage.Memory }

func (storage *TestStorage) FindPilotsByFeatureFlag(ff *rollouts.FeatureFlag) frameless.Iterator {
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
func (storage *TestStorage) FindFlagPilotByExternalPilotID(featureFlagID, externalPilotID string) (*rollouts.Pilot, error) {
	table := storage.TableFor(rollouts.Pilot{})

	for _, v := range table {
		pilot := v.(*rollouts.Pilot)

		if pilot.FeatureFlagID == featureFlagID && pilot.ExternalID == externalPilotID {
			return pilot, nil
		}
	}

	return nil, nil
}

func (storage *TestStorage) FindFlagByName(name string) (*rollouts.FeatureFlag, error) {
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

func (storage *TestStorage) FindTokenByTokenString(tokenStr string) (*security.Token, error) {
	table := storage.TableFor(security.Token{})

	for _, token := range table {
		token := token.(*security.Token)

		if token.Token == tokenStr {
			return token, nil
		}
	}

	return nil, nil
}
