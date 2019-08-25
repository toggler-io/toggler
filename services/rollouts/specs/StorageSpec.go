package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"github.com/toggler-io/toggler/services/rollouts"
)

type StorageSpec struct {
	Subject rollouts.Storage
	specs.FixtureFactory
}

func (spec StorageSpec) Benchmark(b *testing.B) {
	b.Run(`rollouts`, func(b *testing.B) {

		b.Run(`FeatureFlag`, func(b *testing.B) {
			specs.CommonSpec{
				EntityType:     rollouts.FeatureFlag{},
				FixtureFactory: spec.FixtureFactory,
				Subject:        spec.Subject,
			}.Benchmark(b)

			FlagFinderSpec{
				Subject:        spec.Subject,
				FixtureFactory: spec.FixtureFactory,
			}.Benchmark(b)
		})

		b.Run(`Pilot`, func(b *testing.B) {
			flag := spec.FixtureFactory.Create(rollouts.FeatureFlag{}).(*rollouts.FeatureFlag)
			require.Nil(b, spec.Subject.Save(spec.Context(), flag))
			defer func() { require.Nil(b, spec.Subject.DeleteByID(spec.Context(), rollouts.FeatureFlag{}, flag.ID)) }()

			ff := &FixtureFactoryForPilots{
				FixtureFactory: spec.FixtureFactory,
				FlagID:         flag.ID,
			}

			specs.CommonSpec{
				EntityType:     rollouts.Pilot{},
				FixtureFactory: ff,
				Subject:        spec.Subject,
			}.Benchmark(b)

			pilotFinderSpec{
				Subject:        spec.Subject,
				FixtureFactory: ff,
			}.Benchmark(b)
		})

	})
}

func (spec StorageSpec) Test(t *testing.T) {
	t.Run(`rollouts`, func(t *testing.T) {

		t.Run(`FeatureFlag`, func(t *testing.T) {
			specs.CommonSpec{
				EntityType:     rollouts.FeatureFlag{},
				FixtureFactory: spec.FixtureFactory,
				Subject:        spec.Subject,
			}.Test(t)

			FlagFinderSpec{
				Subject:        spec.Subject,
				FixtureFactory: spec.FixtureFactory,
			}.Test(t)

			s := testcase.NewSpec(t)

			s.Context(`name is uniq across storage`, func(s *testcase.Spec) {
				subject := func(t *testcase.T) error {
					return spec.Subject.Save(spec.Context(), t.I(`flag`).(*rollouts.FeatureFlag))
				}

				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.Truncate(spec.Context(), rollouts.FeatureFlag{}))
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

		t.Run(`Pilot`, func(t *testing.T) {

			flag := spec.FixtureFactory.Create(rollouts.FeatureFlag{}).(*rollouts.FeatureFlag)
			require.Nil(t, spec.Subject.Save(spec.Context(), flag))
			defer func() { require.Nil(t, spec.Subject.Truncate(spec.Context(), rollouts.FeatureFlag{})) }()

			ff := &FixtureFactoryForPilots{
				FixtureFactory: spec.FixtureFactory,
				FlagID:         flag.ID,
			}

			specs.CommonSpec{
				EntityType:     rollouts.Pilot{},
				FixtureFactory: ff,
				Subject:        spec.Subject,
			}.Test(t)

			pilotFinderSpec{
				Subject:        spec.Subject,
				FixtureFactory: ff,
			}.Test(t)

		})
	})
}

type FixtureFactoryForPilots struct {
	specs.FixtureFactory
	FlagID string
}

func (ff *FixtureFactoryForPilots) Create(EntityType interface{}) interface{} {
	switch reflects.BaseValueOf(EntityType).Interface().(type) {
	case rollouts.Pilot:
		pilot := ff.FixtureFactory.Create(EntityType).(*rollouts.Pilot)
		pilot.FeatureFlagID = ff.FlagID
		return pilot

	default:
		return ff.FixtureFactory.Create(EntityType)
	}
}
