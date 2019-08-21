package specs

import (
	"context"
	"testing"

	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/stretchr/testify/require"
)

type StorageSpec struct {
	Subject rollouts.Storage

	FixtureFactory interface {
		specs.FixtureFactory
		SetPilotFeatureFlagID(ffID string) func()
	}
}

func (spec StorageSpec) Benchmark(b *testing.B) {
	b.Run(`rollouts.StorageSpec`, func(b *testing.B) {
		b.Skip()
	})
}

func (spec StorageSpec) Test(t *testing.T) {
	t.Run(`rollouts.StorageSpec`, func(t *testing.T) {
		s := testcase.NewSpec(t)

		testEntity := func(t *testing.T, entityType interface{}) {
			specs.MinimumRequirementsSpec{
				EntityType:     entityType,
				FixtureFactory: spec.FixtureFactory,
				Subject:        spec.Subject,
			}.Test(t)

			specs.UpdaterSpec{
				EntityType:     entityType,
				FixtureFactory: spec.FixtureFactory,
				Subject:        spec.Subject,
			}.Test(t)

			specs.FinderSpec{
				EntityType:     entityType,
				FixtureFactory: spec.FixtureFactory,
				Subject:        spec.Subject,
			}.Test(t)
		}

		s.Describe(`flag`, func(s *testcase.Spec) {
			testEntity(t, rollouts.FeatureFlag{})
			FlagFinderSpec{Subject: spec.Subject, FixtureFactory: spec.FixtureFactory}.Test(t)
		})

		s.Describe(`pilot`, func(s *testcase.Spec) {
			s.Let(`flag`, func(t *testcase.T) interface{} {
				return spec.FixtureFactory.Create(rollouts.FeatureFlag{})
			})

			s.Around(func(t *testcase.T) func() {
				flag := t.I(`flag`).(*rollouts.FeatureFlag)
				require.Nil(t, spec.Subject.Save(spec.ctx(), flag))
				td := spec.FixtureFactory.SetPilotFeatureFlagID(flag.ID)
				return func() {
					require.Nil(t, spec.Subject.Truncate(spec.ctx(), rollouts.FeatureFlag{}))
					td()
				}
			})

			s.Then(`coverage pass`, func(t *testcase.T) {
				testEntity(t.T, rollouts.Pilot{})
				PilotFinderSpec{
					Subject:        spec.Subject,
					FixtureFactory: spec.FixtureFactory,
				}.Test(t.T)
			})
		})

		s.Describe(`flag name uniq across storage`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) error {
				return spec.Subject.Save(spec.ctx(), t.I(`flag`).(*rollouts.FeatureFlag))
			}

			s.Before(func(t *testcase.T) {
				require.Nil(t, spec.Subject.Truncate(spec.ctx(), rollouts.FeatureFlag{}))
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

func (spec StorageSpec) ctx() context.Context {
	return spec.FixtureFactory.Context()
}
