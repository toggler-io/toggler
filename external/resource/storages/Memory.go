package storages

import (
	"context"
	"fmt"
	"github.com/adamluzsi/frameless/reflects"

	"github.com/adamluzsi/frameless/inmemory"
	"github.com/adamluzsi/frameless/iterators"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
)

func NewEventLogMemoryStorage() *Memory {
	return &Memory{EventLog: inmemory.NewEventLog()}
}

type Memory struct {
	EventLog *inmemory.EventLog

	Options struct {
		EventLogging bool
	}

	closed bool
}

func (s *Memory) storageFor(T interface{}) *inmemory.EventLogStorage {
	return inmemory.NewEventLogStorageWithNamespace(T, s.EventLog, reflects.SymbolicName(T))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Memory) BeginTx(ctx context.Context) (context.Context, error) {
	return s.EventLog.BeginTx(ctx)
}

func (s *Memory) CommitTx(ctx context.Context) error {
	return s.EventLog.CommitTx(ctx)
}

func (s *Memory) RollbackTx(ctx context.Context) error {
	return s.EventLog.RollbackTx(ctx)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Memory) ReleaseFlag(ctx context.Context) release.FlagStorage {
	return &MemoryReleaseFlagStorage{EventLogStorage: s.storageFor(release.Flag{})}
}

type MemoryReleaseFlagStorage struct {
	*inmemory.EventLogStorage
}

func (s *MemoryReleaseFlagStorage) FindReleaseFlagByName(ctx context.Context, name string) (*release.Flag, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var flag *release.Flag
	for _, v := range s.View(ctx) {
		flagRecord := v.(release.Flag)

		if flagRecord.Name == name {
			f := flagRecord
			flag = &f
			break
		}
	}

	return flag, nil
}

func (s *MemoryReleaseFlagStorage) FindReleaseFlagsByName(ctx context.Context, names ...string) release.FlagEntries {
	if err := ctx.Err(); err != nil {
		return iterators.NewError(err)
	}

	var flags []release.Flag

	nameIndex := make(map[string]struct{})

	for _, name := range names {
		nameIndex[name] = struct{}{}
	}

	for _, v := range s.View(ctx) {
		flag := v.(release.Flag)

		if _, ok := nameIndex[flag.Name]; ok {
			flags = append(flags, flag)
		}
	}

	return iterators.NewSlice(flags)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Memory) ReleasePilot(ctx context.Context) release.PilotStorage {
	return &MemoryReleasePilotStorage{EventLogStorage: s.storageFor(release.ManualPilot{})}
}

type MemoryReleasePilotStorage struct {
	*inmemory.EventLogStorage
}

func (s *MemoryReleasePilotStorage) FindReleaseManualPilotByExternalID(ctx context.Context, flagID, envID interface{}, pilotExtID string) (*release.ManualPilot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var p *release.ManualPilot
	for _, v := range s.View(ctx) {
		pilot := v.(release.ManualPilot)

		if pilot.FlagID == flagID && pilot.DeploymentEnvironmentID == envID && pilot.ExternalID == pilotExtID {
			p = &pilot
			break
		}
	}

	return p, nil
}

func (s *MemoryReleasePilotStorage) FindReleasePilotsByReleaseFlag(ctx context.Context, flag release.Flag) release.PilotEntries {
	if err := ctx.Err(); err != nil {
		return iterators.NewError(err)
	}

	var pilots []release.ManualPilot
	for _, v := range s.View(ctx) {
		pilot := v.(release.ManualPilot)

		if pilot.FlagID == flag.ID {
			pilots = append(pilots, pilot)
		}
	}

	return iterators.NewSlice(pilots)
}

func (s *MemoryReleasePilotStorage) FindReleasePilotsByExternalID(ctx context.Context, pilotPublicID string) release.PilotEntries {
	if err := ctx.Err(); err != nil {
		return iterators.NewError(err)
	}

	var pilots []release.ManualPilot
	for _, v := range s.View(ctx) {
		p := v.(release.ManualPilot)

		if p.ExternalID == pilotPublicID {
			pilots = append(pilots, p)
		}
	}

	return iterators.NewSlice(pilots)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Memory) ReleaseRollout(ctx context.Context) release.RolloutStorage {
	return &MemoryReleaseRolloutStorage{EventLogStorage: s.storageFor(release.Rollout{})}
}

type MemoryReleaseRolloutStorage struct {
	*inmemory.EventLogStorage
}

func (s *MemoryReleaseRolloutStorage) FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx context.Context, flag release.Flag, env deployment.Environment, ptr *release.Rollout) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	var found bool
	for _, v := range s.View(ctx) {
		r := v.(release.Rollout) // copy

		if r.FlagID == flag.ID && r.DeploymentEnvironmentID == env.ID {
			*ptr = r
			found = true
			break
		}
	}

	return found, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Memory) DeploymentEnvironment(ctx context.Context) deployment.EnvironmentStorage {
	return &MemoryDeploymentEnvironmentStorage{EventLogStorage: s.storageFor(deployment.Environment{})}
}

type MemoryDeploymentEnvironmentStorage struct {
	*inmemory.EventLogStorage
}

func (s *MemoryDeploymentEnvironmentStorage) FindDeploymentEnvironmentByAlias(ctx context.Context, idOrName string, env *deployment.Environment) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	var found bool
	for _, v := range s.View(ctx) {
		e := v.(deployment.Environment)

		if e.ID == idOrName || e.Name == idOrName {
			*env = e
			found = true
			break
		}
	}

	return found, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Memory) SecurityToken(ctx context.Context) security.TokenStorage {
	return &MemorySecurityTokenStorage{EventLogStorage: s.storageFor(security.Token{})}
}

type MemorySecurityTokenStorage struct {
	*inmemory.EventLogStorage
}

func (s *MemorySecurityTokenStorage) FindTokenBySHA512Hex(ctx context.Context, tokenAsText string) (*security.Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var token *security.Token

	for _, v := range s.View(ctx) {
		tkn := v.(security.Token)

		if tkn.SHA512 == tokenAsText {
			t := tkn
			token = &t
			break
		}
	}

	return token, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Memory) Close() error {
	if s.closed {
		return fmt.Errorf(`dev storage already closed`)
	}
	s.closed = true
	return nil
}
