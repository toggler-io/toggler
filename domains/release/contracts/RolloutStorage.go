package contracts

import (
	"context"
	"testing"

	"github.com/adamluzsi/frameless"

	"github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	sh "github.com/toggler-io/toggler/spechelper"
)

type RolloutStorage struct {
	Subject        func(testing.TB) release.Storage
	Context        func(testing.TB) context.Context
	FixtureFactory func(testing.TB) frameless.FixtureFactory
}

func (c RolloutStorage) storage() testcase.Var {
	return testcase.Var{
		Name: "release rollout storage",
		Init: func(t *testcase.T) interface{} {
			return sh.StorageGet(t).ReleaseRollout(sh.ContextGet(t))
		},
	}
}

func (c RolloutStorage) storageGet(t *testcase.T) release.RolloutStorage {
	return c.storage().Get(t).(release.RolloutStorage)
}

func (c RolloutStorage) Test(t *testing.T) {
	c.Spec(testcase.NewSpec(t))
}

func (c RolloutStorage) Benchmark(b *testing.B) {
	c.Spec(testcase.NewSpec(b))
}

func (c RolloutStorage) String() string {
	return `RolloutStorage`
}

func (c RolloutStorage) Spec(s *testcase.Spec) {
	sh.Storage.Let(s, func(t *testcase.T) interface{} {
		return c.Subject(t)
	})
	sh.FixtureFactoryLet(s, c.FixtureFactory)

	newRolloutStorage := func(tb testing.TB) release.RolloutStorage {
		return c.Subject(tb).ReleaseRollout(c.Context(tb))
	}

	T := release.Rollout{}
	testcase.RunContract(s,
		contracts.Creator{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return newRolloutStorage(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Finder{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return newRolloutStorage(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Updater{T: T,
			Subject: func(tb testing.TB) contracts.UpdaterSubject {
				return newRolloutStorage(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Deleter{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return newRolloutStorage(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Publisher{T: T,
			Subject: func(tb testing.TB) contracts.PublisherSubject {
				return newRolloutStorage(tb)
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
		contracts.OnePhaseCommitProtocol{T: release.Rollout{},
			Subject: func(tb testing.TB) (frameless.OnePhaseCommitProtocol, contracts.CRD) {
				storage := c.Subject(tb)
				return storage, storage.ReleaseRollout(c.Context(tb))
			},
			Context:        c.Context,
			FixtureFactory: c.FixtureFactory,
		},
	)

	s.Describe(`.FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment`,
		c.specFindReleaseRolloutByReleaseFlagAndDeploymentEnvironment)
}

// TODO replace with FindOne contract
func (c RolloutStorage) specFindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(s *testcase.Spec) {
	var subject = func(t *testcase.T, rollout *release.Rollout) (bool, error) {
		return c.storageGet(t).FindByFlagEnvironment(
			sh.ContextGet(t),
			*sh.ExampleReleaseFlag(t),
			*sh.ExampleDeploymentEnvironment(t),
			rollout,
		)
	}

	const rolloutLetVar = `rollout`

	s.When(`rollout was stored before`, func(s *testcase.Spec) {
		sh.GivenWeHaveReleaseRollout(s,
			rolloutLetVar,
			sh.LetVarExampleReleaseFlag,
			sh.LetVarExampleDeploymentEnvironment,
		)
		s.Before(func(t *testcase.T) { sh.GetReleaseRollout(t, rolloutLetVar) }) // eager load

		s.Then(`it will find the rollout entry`, func(t *testcase.T) {
			var r release.Rollout
			found, err := subject(t, &r)
			require.Nil(t, err)
			require.True(t, found)
			require.Equal(t, *sh.GetReleaseRollout(t, rolloutLetVar), r)
		})
	})

	s.When(`rollout is not in the storage`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			contracts.DeleteAllEntity(t, c.storageGet(t), c.Context(t))
		})

		s.Then(`it will yield no result`, func(t *testcase.T) {
			var r release.Rollout
			found, err := subject(t, &r)
			require.Nil(t, err)
			require.False(t, found)
		})
	})
}
