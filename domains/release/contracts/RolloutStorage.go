package contracts

import (
	"github.com/adamluzsi/frameless"
	"testing"

	"github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	sh "github.com/toggler-io/toggler/spechelper"
)

type RolloutStorage struct {
	Subject        func(testing.TB) release.Storage
	FixtureFactory sh.FixtureFactory
}

func (spec RolloutStorage) storage() testcase.Var {
	return testcase.Var{
		Name: "release rollout storage",
		Init: func(t *testcase.T) interface{} {
			return sh.StorageGet(t).ReleaseRollout(sh.ContextGet(t))
		},
	}
}

func (spec RolloutStorage) storageGet(t *testcase.T) release.RolloutStorage {
	return spec.storage().Get(t).(release.RolloutStorage)
}

func (spec RolloutStorage) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec RolloutStorage) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec RolloutStorage) setUp(s *testcase.Spec) {
	// because we use interactors from the spec_helper in this contract
	// - FixtureFactory.Dynamic
	// - Example...
	sh.SetUp(s)

	sh.Storage.Let(s, func(t *testcase.T) interface{} {
		return spec.Subject(t)
	})
}

func (spec RolloutStorage) Spec(tb testing.TB) {
	testcase.NewSpec(tb).Describe(`RolloutStorage`, func(s *testcase.Spec) {
		spec.setUp(s)

		newRolloutStorage := func(tb testing.TB) release.RolloutStorage {
			return  spec.Subject(tb).ReleaseRollout(spec.FixtureFactory.Context())
		}

		s.Test(`contracts`, func(t *testcase.T) {
			T := release.Rollout{}
			testcase.RunContract(t,
				contracts.Creator{T: T,
					Subject: func(tb testing.TB) contracts.CRD {
						return newRolloutStorage(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
				contracts.Finder{T: T,
					Subject: func(tb testing.TB) contracts.CRD {
						return newRolloutStorage(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
				contracts.Updater{T: T,
					Subject: func(tb testing.TB) contracts.UpdaterSubject {
						return newRolloutStorage(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
				contracts.Deleter{T: T,
					Subject: func(tb testing.TB) contracts.CRD {
						return newRolloutStorage(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
				contracts.Publisher{T: T,
					Subject: func(tb testing.TB) contracts.PublisherSubject {
						return newRolloutStorage(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
				contracts.OnePhaseCommitProtocol{T: release.Rollout{},
					Subject: func(tb testing.TB) (frameless.OnePhaseCommitProtocol, contracts.CRD) {
						storage := spec.Subject(tb)
						return storage, storage.ReleaseRollout(spec.FixtureFactory.Context())
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
			)
		})

		s.Describe(`.FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment`,
			spec.specFindReleaseRolloutByReleaseFlagAndDeploymentEnvironment)
	})
}

// TODO replace with FindOne contract
func (spec RolloutStorage) specFindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(s *testcase.Spec) {
	var subject = func(t *testcase.T, rollout *release.Rollout) (bool, error) {
		return spec.storageGet(t).FindByFlagEnvironment(
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
			contracts.DeleteAllEntity(t, spec.storageGet(t), spec.FixtureFactory.Context())
		})

		s.Then(`it will yield no result`, func(t *testcase.T) {
			var r release.Rollout
			found, err := subject(t, &r)
			require.Nil(t, err)
			require.False(t, found)
		})
	})
}
