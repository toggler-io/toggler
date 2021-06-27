package caches

import (
	"context"
	"fmt"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/cache"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/domains/toggler"
	"sort"
	"strings"
)

type FlagStorage struct {
	*cache.Manager
	Source toggler.Storage
}

func (s *FlagStorage) FindReleaseFlagByName(ctx context.Context, name string) (*release.Flag, error) {
	queryID := fmt.Sprintf("FindReleaseFlagByName/%s", name)
	var flag release.Flag
	found, err := s.Manager.CacheQueryOne(ctx, queryID, &flag, func(ptr interface{}) (found bool, err error) {
		f, err := s.Source.ReleaseFlag(ctx).FindReleaseFlagByName(ctx, name)
		if err != nil {
			return false, err
		}
		if f == nil {
			return false, nil
		}
		return true, reflects.Link(*f, ptr)
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &flag, nil
}

func (s *FlagStorage) FindReleaseFlagsByName(ctx context.Context, names ...string) release.FlagEntries {
	sort.Strings(names)
	queryID := fmt.Sprintf(`FindReleaseFlagsByName/%s`, strings.Join(names, ","))
	return s.Manager.CacheQueryMany(ctx, queryID, func() frameless.Iterator {
		return s.Source.ReleaseFlag(ctx).FindReleaseFlagsByName(ctx, names...)
	})
}

type PilotStorage struct {
	*cache.Manager
	Source toggler.Storage
}

func (s *PilotStorage) FindReleaseManualPilotByExternalID(ctx context.Context, flagID, envID interface{}, pilotExtID string) (*release.Pilot, error) {
	var pilot release.Pilot
	queryID := fmt.Sprintf("FindReleaseManualPilotByExternalID/flag:%v/env:%v/pilotExtID:%s", flagID, envID, pilotExtID)
	found, err := s.Manager.CacheQueryOne(ctx, queryID, &pilot, func(ptr interface{}) (found bool, err error) {
		p, err := s.Source.ReleasePilot(ctx).FindReleaseManualPilotByExternalID(ctx, flagID, envID, pilotExtID)
		if err != nil {
			return false, err
		}
		if p == nil {
			return false, nil
		}
		return true, reflects.Link(*p, ptr)
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &pilot, nil
}

func (s *PilotStorage) FindReleasePilotsByReleaseFlag(ctx context.Context, flag release.Flag) release.PilotEntries {
	return s.Manager.CacheQueryMany(ctx, fmt.Sprintf("FindReleasePilotsByReleaseFlag/flagID:%s", flag.ID), func() frameless.Iterator {
		return s.Source.ReleasePilot(ctx).FindReleasePilotsByReleaseFlag(ctx, flag)
	})
}

func (s *PilotStorage) FindReleasePilotsByExternalID(ctx context.Context, externalID string) release.PilotEntries {
	return s.Manager.CacheQueryMany(ctx, fmt.Sprintf("FindReleasePilotsByExternalID/%s", externalID), func() frameless.Iterator {
		return s.Source.ReleasePilot(ctx).FindReleasePilotsByExternalID(ctx, externalID)
	})
}

type RolloutStorage struct {
	*cache.Manager
	Source toggler.Storage
}

func (s *RolloutStorage) FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx context.Context, flag release.Flag, environment release.Environment, rollout *release.Rollout) (bool, error) {
	queryID := fmt.Sprintf("FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment/flag:%s/env:%s", flag.ID, environment.ID)
	return s.CacheQueryOne(ctx, queryID, rollout, func(ptr interface{}) (found bool, err error) {
		return s.Source.ReleaseRollout(ctx).
			FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx, flag, environment, ptr.(*release.Rollout))
	})
}

type EnvironmentStorage struct {
	*cache.Manager
	Source toggler.Storage
}

func (s *EnvironmentStorage) FindDeploymentEnvironmentByAlias(ctx context.Context, idOrName string, env *release.Environment) (bool, error) {
	s.Source.ReleaseEnvironment(ctx)

	queryID := fmt.Sprintf("FindDeploymentEnvironmentByAlias/%s", idOrName)
	return s.Manager.CacheQueryOne(ctx, queryID, env, func(ptr interface{}) (found bool, err error) {
		return s.Source.ReleaseEnvironment(ctx).FindDeploymentEnvironmentByAlias(ctx, idOrName, ptr.(*release.Environment))
	})
}

type TokenStorage struct {
	*cache.Manager
	Source toggler.Storage
}

func (s *TokenStorage) FindTokenBySHA512Hex(ctx context.Context, sha512hex string) (*security.Token, error) {
	queryID := fmt.Sprintf("FindTokenBySHA512Hex/%s", sha512hex)
	var token security.Token
	found, err := s.Manager.CacheQueryOne(ctx, queryID, &token, func(ptr interface{}) (found bool, err error) {
		t, err := s.Source.SecurityToken(ctx).FindTokenBySHA512Hex(ctx, sha512hex)
		if err != nil {
			return false, err
		}
		if t == nil {
			return false, nil
		}
		return true, reflects.Link(*t, ptr)
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &token, nil
}
