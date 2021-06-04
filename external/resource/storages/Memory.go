package storages

import (
	"context"
	"fmt"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources/storages/inmemory"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
)

func NewMemory() *Memory {
	s := inmemory.New()
	s.Options.DisableEventLogging = true
	return &Memory{Storage: s}
}

type Memory struct {
	*inmemory.Storage

	closed bool
}

func (s *Memory) Close() error {
	if s.closed {
		return fmt.Errorf(`dev storage already closed`)
	}
	s.closed = true
	return nil
}

func (s *Memory) FindReleasePilotsByExternalID(ctx context.Context, pilotExtID string) release.PilotEntries {
	var pilots []release.ManualPilot

	_ = s.Storage.InTx(ctx, func(tx *inmemory.MemoryTransaction) error {
		for _, e := range tx.ViewFor(release.ManualPilot{}) {
			p := e.(release.ManualPilot)

			if p.ExternalID == pilotExtID {
				pilots = append(pilots, p)
			}
		}

		return nil
	})

	return iterators.NewSlice(pilots)
}

func (s *Memory) FindReleaseFlagsByName(ctx context.Context, names ...string) iterators.Interface {
	var flags []release.Flag

	nameIndex := make(map[string]struct{})

	for _, name := range names {
		nameIndex[name] = struct{}{}
	}

	_ = s.InTx(ctx, func(tx *inmemory.MemoryTransaction) error {
		for _, e := range tx.ViewFor(release.Flag{}) {
			flag := e.(release.Flag)

			if _, ok := nameIndex[flag.Name]; ok {
				flags = append(flags, flag)
			}
		}

		return nil
	})

	return iterators.NewSlice(flags)
}

func (s *Memory) FindReleasePilotsByReleaseFlag(ctx context.Context, flag release.Flag) release.PilotEntries {
	var pilots []release.ManualPilot

	_ = s.InTx(ctx, func(tx *inmemory.MemoryTransaction) error {
		for _, v := range tx.ViewFor(release.ManualPilot{}) {
			pilot := v.(release.ManualPilot)

			if pilot.FlagID == flag.ID {
				pilots = append(pilots, pilot)
			}
		}

		return nil
	})

	return iterators.NewSlice(pilots)
}

func (s *Memory) FindReleaseManualPilotByExternalID(ctx context.Context, flagID, envID interface{}, pilotExtID string) (*release.ManualPilot, error) {
	var p *release.ManualPilot

	_ = s.Storage.InTx(ctx, func(tx *inmemory.MemoryTransaction) error {
		for _, v := range tx.ViewFor(release.ManualPilot{}) {
			pilot := v.(release.ManualPilot)

			if pilot.FlagID == flagID && pilot.DeploymentEnvironmentID == envID && pilot.ExternalID == pilotExtID {
				p = &pilot
				return nil
			}
		}

		return nil
	})

	return p, nil
}

func (s *Memory) FindDeploymentEnvironmentByAlias(ctx context.Context, idOrName string, env *deployment.Environment) (bool, error) {
	var found bool
	return found, s.Storage.InTx(ctx, func(tx *inmemory.MemoryTransaction) error {
		for _, v := range tx.ViewFor(deployment.Environment{}) {
			record := v.(deployment.Environment)

			if record.ID == idOrName || record.Name == idOrName {
				*env = record
				found = true
				return nil
			}
		}

		return nil
	})
}

func (s *Memory) FindReleaseFlagByName(ctx context.Context, name string) (*release.Flag, error) {
	var flag *release.Flag

	_ = s.Storage.InTx(ctx, func(tx *inmemory.MemoryTransaction) error {
		for _, v := range tx.ViewFor(release.Flag{}) {
			flagRecord := v.(release.Flag)

			if flagRecord.Name == name {
				f := flagRecord
				flag = &f
				return nil
			}
		}

		return nil
	})

	return flag, nil
}

func (s *Memory) FindTokenBySHA512Hex(ctx context.Context, tokenAsText string) (*security.Token, error) {
	var token *security.Token

	_ = s.Storage.InTx(ctx, func(tx *inmemory.MemoryTransaction) error {
		for _, tkn := range tx.ViewFor(security.Token{}) {
			tkn := tkn.(security.Token)

			if tkn.SHA512 == tokenAsText {
				t := tkn
				token = &t
				return nil
			}
		}

		return nil
	})
	return token, nil
}

func (s *Memory) FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(ctx context.Context, flag release.Flag, env deployment.Environment, ptr *release.Rollout) (bool, error) {
	var found bool

	_ = s.Storage.InTx(ctx, func(tx *inmemory.MemoryTransaction) error {
		for _, rollout := range tx.ViewFor(release.Rollout{}) {
			rollout := rollout.(release.Rollout)

			if rollout.FlagID == flag.ID && rollout.DeploymentEnvironmentID == env.ID {
				*ptr = rollout
				found = true
				return nil
			}
		}

		return nil
	})

	return found, nil
}