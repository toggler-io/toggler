package caches_test

import (
	"context"
	"testing"

	"github.com/adamluzsi/frameless"
	csh "github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/testcase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/domains/toggler/contracts"
	"github.com/toggler-io/toggler/external/resource/caches"
	"github.com/toggler-io/toggler/external/resource/storages"
	sh "github.com/toggler-io/toggler/spechelper"
)

var (
	_ toggler.Storage            = &caches.Memory{}
	_ release.Storage            = &caches.Memory{}
	_ security.Storage           = &caches.Memory{}
	_ release.EnvironmentStorage = &caches.EnvironmentStorage{}
	_ release.FlagStorage        = &caches.FlagStorage{}
	_ release.RolloutStorage     = &caches.RolloutStorage{}
	_ release.PilotStorage       = &caches.PilotStorage{}
)

func TestMemory_smoke(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Test(``, func(t *testcase.T) {
		storage := storages.NewEventLogMemoryStorage()
		storage.EventLog.Options.DisableAsyncSubscriptionHandling = true
		m, err := caches.NewMemory(storage)
		require.Nil(t, err)
		t.Cleanup(func() { require.Nil(t, m.Close()) })
		ff := sh.NewFixtureFactory(t)
		ctx := context.Background()
		ts := m.SecurityToken(ctx)
		token := ff.Fixture(security.Token{}, ctx).(security.Token)
		csh.CreateEntity(t, ts, ctx, &token)
		token.OwnerUID = uuid.New().String()
		csh.UpdateEntity(t, ts, ctx, &token)
	})
}

func TestMemory(t *testing.T)      { SpecMemory(t) }
func BenchmarkMemory(b *testing.B) { SpecMemory(b) }

func SpecMemory(tb testing.TB) {
	testcase.RunContract(sh.NewSpec(tb), contracts.Storage{
		Subject: func(tb testing.TB) toggler.Storage {
			storage := storages.NewEventLogMemoryStorage()
			storage.EventLog.Options.DisableAsyncSubscriptionHandling = true
			m, err := caches.NewMemory(storage)
			require.Nil(tb, err)
			tb.Cleanup(func() { require.Nil(tb, m.Close()) })
			return m
		},
		FixtureFactory: func(tb testing.TB) frameless.FixtureFactory {
			return sh.NewFixtureFactory(tb)
		},
		Context: func(tb testing.TB) context.Context {
			return context.Background()
		},
	})
}
