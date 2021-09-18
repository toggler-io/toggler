package spechelper

import (
	"os"

	"github.com/adamluzsi/frameless/inmemory"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

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
		s.Options.EventLogging = true
		s.EventLog.Options.DisableAsyncSubscriptionHandling = true
		inmemory.LogHistoryOnFailure(t, s.EventLog)
		t.Defer(s.Close)
		return s
	}

	storage, err := storages.New(connstr)
	require.Nil(t, err)
	t.Defer(storage.Close)

	var cleanup = func() {
		require.Nil(t, storage.SecurityToken(ContextGet(t)).DeleteAll(ContextGet(t)))
		require.Nil(t, storage.ReleasePilot(ContextGet(t)).DeleteAll(ContextGet(t)))
		require.Nil(t, storage.ReleaseRollout(ContextGet(t)).DeleteAll(ContextGet(t)))
		require.Nil(t, storage.ReleaseFlag(ContextGet(t)).DeleteAll(ContextGet(t)))
		require.Nil(t, storage.ReleaseEnvironment(ContextGet(t)).DeleteAll(ContextGet(t)))
	}

	// TODO: replace this solution for external interface testing with middleware approach where tx is injected to the request context.
	// 	go runs each package tests in parallel, and as soon there would be multiple external interface tests, it would cause side effects between the two package testing suite.
	if !t.HasTag(TagBlackBox) {
		ctx, err := storage.BeginTx(ContextGet(t))
		require.Nil(t, err)
		Context.Set(t, ctx)
		t.Defer(storage.RollbackTx, ctx)
	} else {
		// cleanup
		t.Defer(cleanup)
	}

	// clean ahead
	cleanup()

	return storage
}
