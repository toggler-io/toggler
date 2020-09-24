package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	. "github.com/toggler-io/toggler/testing"
)

type RolloutStorage struct {
	Subject interface {
		release.RolloutFinder
		resources.Creator
		resources.Finder
		resources.Deleter
		resources.Updater
		resources.OnePhaseCommitProtocol
		resources.CreatorPublisher
		resources.UpdaterPublisher
		resources.DeleterPublisher
	}

	FixtureFactory FixtureFactory
}

func (spec RolloutStorage) Test(t *testing.T) {
	t.Run(`RolloutStorage`, func(t *testing.T) {
		spec.Spec(t)
	})
}

func (spec RolloutStorage) setup(s *testcase.Spec) {
	SetUp(s)

	s.Let(LetVarExampleStorage, func(t *testcase.T) interface{} {
		return spec.Subject
	})

	s.Before(func(t *testcase.T) {
		require.Nil(t, spec.Subject.DeleteAll(GetContext(t), deployment.Environment{}))
		require.Nil(t, spec.Subject.DeleteAll(GetContext(t), release.Flag{}))
		require.Nil(t, spec.Subject.DeleteAll(GetContext(t), release.Rollout{}))
	})
}

func (spec RolloutStorage) Benchmark(b *testing.B) {
	b.Run(`RolloutStorage`, func(b *testing.B) {
		spec.Spec(b)
	})
}

func (spec RolloutStorage) Spec(tb testing.TB) {
	s := testcase.NewSpec(tb)
	spec.setup(s)

	s.Test(``, func(t *testcase.T) {
		specs.Run(t.T,
			specs.OnePhaseCommitProtocol{
				EntityType:     release.Rollout{},
				Subject:        spec.Subject,
				FixtureFactory: spec.FixtureFactory.Dynamic(t),
			},
			specs.CRUD{
				EntityType:     release.Rollout{},
				FixtureFactory: spec.FixtureFactory.Dynamic(t),
				Subject:        spec.Subject,
			},
			specs.CreatorPublisher{
				Subject:        spec.Subject,
				EntityType:     release.Rollout{},
				FixtureFactory: spec.FixtureFactory.Dynamic(t),
			},
			specs.UpdaterPublisher{
				Subject:        spec.Subject,
				EntityType:     release.Rollout{},
				FixtureFactory: spec.FixtureFactory.Dynamic(t),
			},
			specs.DeleterPublisher{
				Subject:        spec.Subject,
				EntityType:     release.Rollout{},
				FixtureFactory: spec.FixtureFactory.Dynamic(t),
			},
		)
	})

	s.Describe(`#FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment`,
		spec.specFindReleaseRolloutByReleaseFlagAndDeploymentEnvironment)
}

func (spec RolloutStorage) specFindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(s *testcase.Spec) {
	var subject = func(t *testcase.T, rollout *release.Rollout) (bool, error) {
		return spec.Subject.FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(
			GetContext(t),
			*ExampleReleaseFlag(t),
			*ExampleDeploymentEnvironment(t),
			rollout,
		)
	}

	const rolloutLetVar = `rollout`

	s.When(`rollout was stored before`, func(s *testcase.Spec) {
		GivenWeHaveReleaseRollout(s,
			rolloutLetVar,
			LetVarExampleReleaseFlag,
			LetVarExampleDeploymentEnvironment,
		)
		s.Before(func(t *testcase.T) { GetReleaseRollout(t, rolloutLetVar) }) // eager load

		s.Then(`it will find the rollout entry`, func(t *testcase.T) {
			var r release.Rollout
			found, err := subject(t, &r)
			require.Nil(t, err)
			require.True(t, found)
			require.Equal(t, *GetReleaseRollout(t, rolloutLetVar), r)
		})
	})

	s.When(`rollout is not in the storage`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, spec.Subject.DeleteAll(GetContext(t), release.Rollout{}))
		})

		s.Then(`it will yield no result`, func(t *testcase.T) {
			var r release.Rollout
			found, err := subject(t, &r)
			require.Nil(t, err)
			require.False(t, found)
		})
	})
}
