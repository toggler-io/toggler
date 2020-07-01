package specs

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	. "github.com/toggler-io/toggler/testing"
)

type RolloutStorageSpec struct {
	Subject interface {
		release.RolloutFinder
		resources.Creator
		resources.Finder
		resources.Deleter
		resources.Updater
		resources.OnePhaseCommitProtocol
	}
	specs.FixtureFactory
}

func (spec RolloutStorageSpec) Test(t *testing.T) {
	s := testcase.NewSpec(t)
	SetUp(s)

	s.Let(ExampleStorageLetVar, func(t *testcase.T) interface{} {
		return spec.Subject
	})

	s.Before(func(t *testcase.T) {
		require.Nil(t, spec.Subject.DeleteAll(GetContext(t), deployment.Environment{}))
		require.Nil(t, spec.Subject.DeleteAll(GetContext(t), release.Flag{}))
		require.Nil(t, spec.Subject.DeleteAll(GetContext(t), release.Rollout{}))
	})

	s.Context(`RolloutStorageSpec`, func(s *testcase.Spec) {
		s.Test(`CommonSpec#Rollout`, func(t *testcase.T) {
			specs.CommonSpec{
				EntityType: release.Rollout{},
				FixtureFactory: RolloutStorageSpecFixtureFactory{
					FixtureFactory: spec.FixtureFactory,
					flag:           ExampleReleaseFlag(t),
					env:            ExampleDeploymentEnvironment(t),
				},
				Subject: spec.Subject,
			}.Test(t.T)
		})

		s.Test(`OnePhaseCommitProtocol`, func(t *testcase.T) {
			specs.OnePhaseCommitProtocolSpec{
				EntityType: release.Rollout{},
				Subject:    spec.Subject,
				FixtureFactory: RolloutStorageSpecFixtureFactory{
					FixtureFactory: spec.FixtureFactory,
					flag:           ExampleReleaseFlag(t),
					env:            ExampleDeploymentEnvironment(t),
				},
			}.Test(t.T)
		})

		s.Describe(`#FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment`, func(s *testcase.Spec) {
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
					ExampleReleaseFlagLetVar,
					ExampleDeploymentEnvironmentLetVar,
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
		})
	})
}

func (spec RolloutStorageSpec) Benchmark(b *testing.B) {
	b.Skip()
}

type RolloutStorageSpecFixtureFactory struct {
	specs.FixtureFactory
	env  *deployment.Environment
	flag *release.Flag
}

func (ff RolloutStorageSpecFixtureFactory) Create(EntityType interface{}) (StructPTR interface{}) {
	switch EntityType.(type) {
	case release.Rollout:
		r := ff.FixtureFactory.Create(EntityType).(*release.Rollout)
		r.DeploymentEnvironmentID = ff.env.ID
		r.FlagID = ff.flag.ID
		r.Plan = func() release.RolloutDefinition {
			switch fixtures.Random.IntN(3) {
			case 0:
				byPercentage := release.NewRolloutDecisionByPercentage()
				byPercentage.Percentage = fixtures.Random.IntBetween(0, 100)
				return byPercentage

			case 1:
				byAPI := release.NewRolloutDecisionByAPI()
				u, err := url.ParseRequestURI(fmt.Sprintf(`https://example.com/%s`, url.PathEscape(fixtures.Random.String())))
				if err != nil {
					panic(err.Error())
				}
				byAPI.URL = u
				return byAPI

			case 2:
				byPercentage := release.NewRolloutDecisionByPercentage()
				byPercentage.Percentage = fixtures.Random.IntBetween(0, 100)

				byAPI := release.NewRolloutDecisionByAPI()
				u, err := url.ParseRequestURI(fmt.Sprintf(`https://example.com/%s`, url.PathEscape(fixtures.Random.String())))
				if err != nil {
					panic(err.Error())
				}
				byAPI.URL = u

				return release.RolloutDecisionAND{
					Left:  byPercentage,
					Right: byAPI,
				}

			default:
				panic(`shouldn't be the case`)
			}
		}()
		return r

	default:
		return ff.FixtureFactory.Create(EntityType)
	}
}
