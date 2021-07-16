package storages_test

import (
	"context"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	fc "github.com/adamluzsi/frameless/contracts"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/domains/toggler/contracts"
	"github.com/toggler-io/toggler/external/resource/storages"
	sh "github.com/toggler-io/toggler/spechelper"
)

var (
	_ toggler.Storage            = &storages.Memory{}
	_ release.Storage            = &storages.Memory{}
	_ security.Storage           = &storages.Memory{}
	_ release.EnvironmentStorage = &storages.MemoryReleaseEnvironmentStorage{}
	_ release.FlagStorage        = &storages.MemoryReleaseFlagStorage{}
	_ release.RolloutStorage     = &storages.MemoryReleaseRolloutStorage{}
	_ release.PilotStorage       = &storages.MemoryReleasePilotStorage{}
)

func TestMemory(t *testing.T) {
	SpecMemory(t)
}

func BenchmarkMemory(b *testing.B) {
	SpecMemory(b)
}

func SpecMemory(tb testing.TB) {
	testcase.RunContract(sh.NewSpec(tb), contracts.Storage{
		Subject: func(tb testing.TB) toggler.Storage {
			return storages.NewEventLogMemoryStorage()
		},
		FixtureFactory: func(tb testing.TB) fc.FixtureFactory {
			return sh.NewFixtureFactory(tb)
		},
	})
}

func TestMemoryReleasePilotStorage_smokeTest(t *testing.T) {
	storage := storages.NewEventLogMemoryStorage()
	storage.EventLog.Options.DisableAsyncSubscriptionHandling = true
	t.Cleanup(func() { require.Nil(t, storage.Close()) })

	var (
		ctx                  = context.Background()
		releaseFlagStorage   = storage.ReleaseFlag(ctx)
		deploymentEnvStorage = storage.ReleaseEnvironment(ctx)
		releasePilotStorage  = storage.ReleasePilot(ctx)
		env                  = &release.Environment{Name: fixtures.Random.String()}
		flag                 = &release.Flag{Name: fixtures.Random.String()}
	)

	var pilotCreateEvents []release.Pilot

	sub, err := releasePilotStorage.Subscribe(ctx, StubSub{
		HandleFunc: func(ctx context.Context, event interface{}) error {
			switch event := event.(type) {
			case frameless.EventCreate:
				ent := event.Entity.(release.Pilot)
				pilotCreateEvents = append(pilotCreateEvents, ent)
			}
			return nil
		},
		ErrorFunc: func(ctx context.Context, err error) error {
			t.Fatalf(`%v`, err)
			return nil
		},
	})
	require.Nil(t, err)
	t.Cleanup(func() { require.Nil(t, sub.Close()) })

	fc.CreateEntity(t, deploymentEnvStorage, ctx, env)
	fc.CreateEntity(t, releaseFlagStorage, ctx, flag)
	require.Empty(t, pilotCreateEvents)

	pilot := &release.Pilot{
		FlagID:          flag.ID,
		EnvironmentID:   env.ID,
		PublicID:        "42",
		IsParticipating: true,
	}

	require.Nil(t, releasePilotStorage.Create(ctx, pilot))

	retry := testcase.Retry{Strategy: testcase.Waiter{WaitTimeout: time.Second}}
	retry.Assert(t, func(tb testing.TB) {
		require.Contains(tb, pilotCreateEvents, *pilot)
	})
}

type StubSub struct {
	HandleFunc func(ctx context.Context, event interface{}) error
	ErrorFunc  func(ctx context.Context, err error) error
}

func (sub StubSub) Handle(ctx context.Context, ent interface{}) error {
	return sub.HandleFunc(ctx, ent)
}
func (sub StubSub) Error(ctx context.Context, err error) error {
	return sub.ErrorFunc(ctx, err)
}
