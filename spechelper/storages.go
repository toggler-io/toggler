package spechelper

import (
	"os"
	"sync"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/resource/storages"
)

var Storage = testcase.Var /* (type toggler.Storage) */ {
	Name:  `toggler.Storage`,
	Init:  storageInit,
	OnLet: storageOnLet,
}

func StorageGet(t *testcase.T) toggler.Storage {
	return Storage.Get(t).(toggler.Storage)
}

func storageOnLet(s *testcase.Spec) {
	if _, ok := os.LookupEnv(`TEST_DATABASE_URL`); ok {
		s.HasSideEffect()
		s.Sequential()
		s.Tag(`database`)
	}
}

func storageInit(t *testcase.T) interface{} {
	connstr, ok := os.LookupEnv(`TEST_DATABASE_URL`)
	if !ok { // use fake implementation
		s := storages.NewMemory()
		s.Options.DisableEventLogging = false
		s.Options.DisableAsyncSubscriptionHandling = true

		t.Defer(s.Close)
		t.Defer(func() {
			if t.Failed() {
				s.LogContextHistory(t, GetContext(t))
			}
		})

		return s
	}

	initCachedStorages.Do(func() {
		s, err := storages.New(connstr)
		require.Nil(t, err)
		//t.Defer(storage.Close)
		cachedStorages = s
	})
	var storage = cachedStorages

	// TODO: replace this solution for external interface testing with middleware approach where tx is injected to the request context.
	// 	go runs each package tests in parallel, and as soon there would be multiple external interface tests, it would cause side effects between the two package testing suite.
	if !t.HasTag(TagBlackBox) {
		ctx, err := storage.BeginTx(GetContext(t))
		require.Nil(t, err)
		t.Let(LetVarContext, ctx)
		t.Defer(storage.RollbackTx, ctx)
	}

	{ // TODO: check if this can be removed
		var cleanup = func() {
			require.Nil(t, storage.DeleteAll(GetContext(t), security.Token{}))
			require.Nil(t, storage.DeleteAll(GetContext(t), release.ManualPilot{}))
			require.Nil(t, storage.DeleteAll(GetContext(t), release.Rollout{}))
			require.Nil(t, storage.DeleteAll(GetContext(t), release.Flag{}))
			require.Nil(t, storage.DeleteAll(GetContext(t), deployment.Environment{}))
		}

		cleanup()
		t.Defer(cleanup)
	}

	return storage
}

var (
	cachedStorages     toggler.Storage
	initCachedStorages sync.Once
)
