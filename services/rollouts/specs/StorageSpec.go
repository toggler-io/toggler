package specs

import (
	"context"
	"github.com/adamluzsi/testcase"
	testing2 "github.com/adamluzsi/toggler/testing"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/toggler/services/rollouts"
)

type StorageSpec struct {
	Storage rollouts.Storage

	FixtureFactory interface {
		specs.FixtureFactory
		SetPilotFeatureFlagID(ffID string) func()
	}
}

func (spec StorageSpec) Test(t *testing.T) {
	s := testcase.NewSpec(t)
	ff := testing2.NewFixtureFactory()
	testEntity := func(t *testing.T, entityType interface{}) {
		specs.TestMinimumRequirements(t, spec.Storage, entityType, ff)
		specs.TestUpdate(t, spec.Storage, entityType, ff)
		specs.TestFindAll(t, spec.Storage, entityType, ff)
	}

	s.Describe(`rollouts.StorageSpec`, func(s *testcase.Spec) {
		s.Describe(`flag`, func(s *testcase.Spec) {
			testEntity(t, rollouts.FeatureFlag{})
			FlagFinderSpec{Subject: spec.Storage, FixtureFactory: spec.FixtureFactory}.Test(t)
		})

		s.Describe(`pilot`, func(s *testcase.Spec) {
			s.Let(`flag`, func(t *testcase.T) interface{} {
				return ff.Create(rollouts.FeatureFlag{})
			})

			s.Around(func(t *testcase.T) func() {
				flag := t.I(`flag`).(*rollouts.FeatureFlag)
				require.Nil(t, spec.Storage.Save(spec.ctx(flag), flag))
				td := ff.SetPilotFeatureFlagID(flag.ID)
				return func() {
					require.Nil(t, spec.Storage.Truncate(spec.ctx(rollouts.FeatureFlag{}), rollouts.FeatureFlag{}))
					td()
				}
			})

			s.Then(`coverage pass`, func(t *testcase.T) {
				testEntity(t.T, rollouts.Pilot{})
				PilotFinderSpec{
					Subject:        spec.Storage,
					FixtureFactory: spec.FixtureFactory,
				}.Test(t.T)
			})
		})

		s.Describe(`flag name uniq across storage`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) error {
				return spec.Storage.Save(spec.ctx(rollouts.FeatureFlag{}), t.I(`flag`).(*rollouts.FeatureFlag))
			}

			s.Before(func(t *testcase.T) {
				require.Nil(t, spec.Storage.Truncate(spec.ctx(rollouts.FeatureFlag{}), rollouts.FeatureFlag{}))
			})

			s.Let(`flag`, func(t *testcase.T) interface{} {
				return &rollouts.FeatureFlag{
					Name: `my-uniq-flag-name`,
				}
			})

			s.When(`flag already stored`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) { require.Nil(t, subject(t)) })

				s.Then(`saving again will create error`, func(t *testcase.T) {
					require.Error(t, subject(t))
				})
			})
		})
	})
}

func (spec StorageSpec) ctx(e interface{}) context.Context {
	return spec.FixtureFactory.Context(e)
}
