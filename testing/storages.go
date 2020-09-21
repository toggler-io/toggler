package testing

import (
	"os"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/external/resource/storages"
	"github.com/toggler-io/toggler/usecases"
)

const LetVarExampleStorage = `TestStorage`

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		if connstr, ok := os.LookupEnv(`TEST_DATABASE_URL`); ok {
			s.HasSideEffect()
			s.Sequential()

			s.Let(LetVarExampleStorage, func(t *testcase.T) interface{} {
				storage, err := storages.New(connstr)
				require.Nil(t, err)
				t.Defer(storage.Close)

				if !t.HasTag(TagBlackBox) {
					ctx, err := storage.BeginTx(GetContext(t))
					require.Nil(t, err)
					t.Let(LetVarContext, ctx)
					t.Defer(storage.RollbackTx, ctx)
				}

				{
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
			})
		} else {
			s.Let(LetVarExampleStorage, func(t *testcase.T) interface{} {
				s := storages.NewTestingStorage()
				t.Defer(s.Close)
				t.Defer(func() {
					if t.T != nil && t.T.Failed() {
						s.History().LogWith(t) // print out storage event history
					}
				})

				return s
			})
		}
	})
}

func GetStorage(t *testcase.T, varName string) usecases.Storage {
	return t.I(varName).(usecases.Storage)
}

func ExampleStorage(t *testcase.T) usecases.Storage {
	return GetStorage(t, LetVarExampleStorage)
}
