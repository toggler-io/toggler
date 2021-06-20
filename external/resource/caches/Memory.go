package caches

import (
	"context"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/cache"
	"github.com/adamluzsi/frameless/inmemory"
	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/domains/toggler"
	"sync"
)

func NewMemory(s toggler.Storage) (*Memory, error) {
	m := &Memory{Source: s, Memory: inmemory.NewMemory()}
	return m, m.Init(context.Background())
}

type Memory struct {
	Source toggler.Storage
	Memory *inmemory.Memory

	init                  sync.Once
	releaseFlag           *cache.Manager
	releaseRollout        *cache.Manager
	releasePilot          *cache.Manager
	deploymentEnvironment *cache.Manager
	securityToken         *cache.Manager
}

func (m *Memory) Init(ctx context.Context) error {
	var err error
	m.init.Do(func() {
		newManager := func(T frameless.T, s cache.Source) (*cache.Manager, error) {
			return cache.NewManager(T, newMemoryStorage(T, m.Memory), s)
		}

		m.releaseFlag, err = newManager(release.Flag{}, m.Source.ReleaseFlag(ctx))
		if err != nil {
			return
		}
		m.releaseRollout, err = newManager(release.Rollout{}, m.Source.ReleaseRollout(ctx))
		if err != nil {
			return
		}
		m.releasePilot, err = newManager(release.ManualPilot{}, m.Source.ReleasePilot(ctx))
		if err != nil {
			return
		}
		m.deploymentEnvironment, err = newManager(deployment.Environment{}, m.Source.DeploymentEnvironment(ctx))
		if err != nil {
			return
		}
		m.securityToken, err = newManager(security.Token{}, m.Source.SecurityToken(ctx))
		if err != nil {
			return
		}
	})
	return err
}

func (m *Memory) BeginTx(ctx context.Context) (context.Context, error) {
	ctx, err := m.Source.BeginTx(ctx)
	if err != nil {
		return ctx, err
	}
	return m.Memory.BeginTx(ctx)
}

func (m *Memory) CommitTx(ctx context.Context) error {
	if err := m.Source.CommitTx(ctx); err != nil {
		return err
	}
	return m.Memory.CommitTx(ctx)
}

func (m *Memory) RollbackTx(ctx context.Context) error {
	if err := m.Source.RollbackTx(ctx); err != nil {
		return err
	}
	return m.Memory.RollbackTx(ctx)
}

func (m *Memory) ReleaseFlag(ctx context.Context) release.FlagStorage {
	return &FlagStorage{
		Manager: m.releaseFlag,
		Source:  m.Source,
	}
}

func (m *Memory) ReleasePilot(ctx context.Context) release.PilotStorage {
	return &PilotStorage{
		Manager: m.releasePilot,
		Source:  m.Source,
	}
}

func (m *Memory) ReleaseRollout(ctx context.Context) release.RolloutStorage {
	return &RolloutStorage{
		Manager: m.releaseRollout,
		Source:  m.Source,
	}
}

func (m *Memory) DeploymentEnvironment(ctx context.Context) deployment.EnvironmentStorage {
	return &EnvironmentStorage{
		Manager: m.deploymentEnvironment,
		Source:  m.Source,
	}
}

func (m *Memory) SecurityToken(ctx context.Context) security.TokenStorage {
	return &TokenStorage{
		Manager: m.securityToken,
		Source:  m.Source,
	}
}

func (m *Memory) Close() error {
	_ = m.releaseFlag.Close()
	_ = m.releaseRollout.Close()
	_ = m.releasePilot.Close()
	_ = m.deploymentEnvironment.Close()
	_ = m.securityToken.Close()
	return m.Source.Close()
}

func newMemoryStorage(T frameless.T, m *inmemory.Memory) *storage {
	return &storage{
		OnePhaseCommitProtocol: m,
		HitStorage:             inmemory.NewStorage(cache.Hit{}, m),
		EntityStorage:          inmemory.NewStorage(T, m),
	}
}

type storage struct {
	frameless.OnePhaseCommitProtocol
	HitStorage    cache.HitStorage
	EntityStorage cache.EntityStorage
}

func (s *storage) CacheEntity(ctx context.Context) cache.EntityStorage {
	return s.EntityStorage
}

func (s *storage) CacheHit(ctx context.Context) cache.HitStorage {
	return s.HitStorage
}
