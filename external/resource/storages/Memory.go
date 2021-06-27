package storages

import (
	"context"
	"fmt"
	"github.com/adamluzsi/frameless/reflects"

	"github.com/adamluzsi/frameless/inmemory"
	"github.com/adamluzsi/frameless/iterators"


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

func (s *MemoryReleaseFlagStorage) FindByName(ctx context.Context, name string) (*release.Flag, error) {
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

func (s *MemoryReleaseFlagStorage) FindByNames(ctx context.Context, names ...string) release.FlagEntries {
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
	return &MemoryReleasePilotStorage{EventLogStorage: s.storageFor(release.Pilot{})}
}

type MemoryReleasePilotStorage struct {
	*inmemory.EventLogStorage
}

func (s *MemoryReleasePilotStorage) FindByFlagEnvPublicID(ctx context.Context, flagID, envID interface{}, pilotExtID string) (*release.Pilot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var p *release.Pilot
	for _, v := range s.View(ctx) {
		pilot := v.(release.Pilot)

		if pilot.FlagID == flagID && pilot.EnvironmentID == envID && pilot.PublicID == pilotExtID {
			p = &pilot
			break
		}
	}

	return p, nil
}

func (s *MemoryReleasePilotStorage) FindByFlag(ctx context.Context, flag release.Flag) release.PilotEntries {
	if err := ctx.Err(); err != nil {
		return iterators.NewError(err)
	}

	var pilots []release.Pilot
	for _, v := range s.View(ctx) {
		pilot := v.(release.Pilot)

		if pilot.FlagID == flag.ID {
			pilots = append(pilots, pilot)
		}
	}

	return iterators.NewSlice(pilots)
}

func (s *MemoryReleasePilotStorage) FindByPublicID(ctx context.Context, pilotPublicID string) release.PilotEntries {
	if err := ctx.Err(); err != nil {
		return iterators.NewError(err)
	}

	var pilots []release.Pilot
	for _, v := range s.View(ctx) {
		p := v.(release.Pilot)

		if p.PublicID == pilotPublicID {
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

func (s *MemoryReleaseRolloutStorage) FindByFlagEnvironment(ctx context.Context, flag release.Flag, env release.Environment, ptr *release.Rollout) (bool, error) {
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

func (s *Memory) ReleaseEnvironment(ctx context.Context) release.EnvironmentStorage {
	return &MemoryReleaseEnvironmentStorage{EventLogStorage: s.storageFor(release.Environment{})}
}

type MemoryReleaseEnvironmentStorage struct {
	*inmemory.EventLogStorage
}

func (s *MemoryReleaseEnvironmentStorage) FindByAlias(ctx context.Context, idOrName string, env *release.Environment) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	var found bool
	for _, v := range s.View(ctx) {
		e := v.(release.Environment)

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
