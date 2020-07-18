package storages

import (
	"context"
	"fmt"

	"github.com/adamluzsi/frameless/dev"
	"github.com/adamluzsi/frameless/iterators"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
)

func NewDevStorage() *DevStorage {
	return &DevStorage{Storage: dev.NewStorage()}
}

type DevStorage struct {
	*dev.Storage

	closed bool
}

func (s *DevStorage) FindReleaseFlagByName(ctx context.Context, name string) (*release.Flag, error) {
	iter := s.FindAll(ctx, release.Flag{})
	defer iter.Close()

	for iter.Next() {
		var flag release.Flag
		if err := iter.Decode(&flag); err != nil {
			return nil, err
		}

		if flag.Name == name {
			return &flag, nil
		}
	}

	return nil, iter.Err()
}

func (s *DevStorage) FindReleaseFlagsByName(ctx context.Context, names ...string) release.FlagEntries {
	var flags []release.Flag

	for _, name := range names {
		f, err := s.FindReleaseFlagByName(ctx, name)
		if err != nil {
			return iterators.NewError(err)
		}
		if f != nil {
			flags = append(flags, *f)
		}
	}

	return iterators.NewSlice(flags)
}

func (s *DevStorage) FindReleaseManualPilotByExternalID(ctx context.Context, flagID, envID, pilotExtID string) (*release.ManualPilot, error) {
	iter := s.FindAll(ctx, release.ManualPilot{})
	defer iter.Close()

	for iter.Next() {
		var mp release.ManualPilot
		if err := iter.Decode(&mp); err != nil {
			return nil, err
		}
		if mp.FlagID == flagID && mp.DeploymentEnvironmentID == envID && mp.ExternalID == pilotExtID {
			return &mp, nil
		}
	}

	return nil, iter.Err()
}

func (s *DevStorage) FindReleasePilotsByReleaseFlag(ctx context.Context, flag release.Flag) release.PilotEntries {
	return iterators.Filter(s.FindAll(ctx, release.ManualPilot{}), func(p release.ManualPilot) bool {
		return p.FlagID == flag.ID
	})
}

func (s *DevStorage) FindReleasePilotsByExternalID(ctx context.Context, externalID string) release.PilotEntries {
	return iterators.Filter(s.FindAll(ctx, release.ManualPilot{}), func(p release.ManualPilot) bool {
		return p.ExternalID == externalID
	})
}

func (s *DevStorage) FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx context.Context, flag release.Flag, environment deployment.Environment, rollout *release.Rollout) (bool, error) {
	return iterators.First(iterators.Filter(s.FindAll(ctx, release.Rollout{}), func(r release.Rollout) bool {
		return r.DeploymentEnvironmentID == environment.ID && r.FlagID == flag.ID
	}), rollout)
}

func (s *DevStorage) FindTokenBySHA512Hex(ctx context.Context, sha512hex string) (*security.Token, error) {
	var tkn security.Token

	found, err := iterators.First(iterators.Filter(s.FindAll(ctx, security.Token{}), func(t security.Token) bool {
		return t.SHA512 == sha512hex
	}), &tkn)

	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	return &tkn, nil
}

func (s *DevStorage) FindDeploymentEnvironmentByAlias(ctx context.Context, idOrName string, env *deployment.Environment) (bool, error) {
	return iterators.First(iterators.Filter(s.FindAll(ctx, deployment.Environment{}), func(d deployment.Environment) bool {
		return d.ID == idOrName || d.Name == idOrName
	}), env)
}

func (s *DevStorage) Close() error {
	if s.closed {
		return fmt.Errorf(`dev storage already closed`)
	}
	s.closed = true
	return nil
}
