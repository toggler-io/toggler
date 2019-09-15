package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"github.com/toggler-io/toggler/services/release"
)

type StorageSpec struct {
	Subject release.Storage
	specs.FixtureFactory
}

func (spec StorageSpec) Benchmark(b *testing.B) {
	b.Run(`rollouts`, func(b *testing.B) {

		b.Run(`ReleaseFlag`, func(b *testing.B) {
			specs.CommonSpec{
				EntityType:     release.Flag{},
				FixtureFactory: spec.FixtureFactory,
				Subject:        spec.Subject,
			}.Benchmark(b)

			FlagFinderSpec{
				Subject:        spec.Subject,
				FixtureFactory: spec.FixtureFactory,
			}.Benchmark(b)
		})

		b.Run(`Pilot`, func(b *testing.B) {
			flag := spec.FixtureFactory.Create(release.Flag{}).(*release.Flag)
			require.Nil(b, spec.Subject.Save(spec.Context(), flag))
			defer func() { require.Nil(b, spec.Subject.DeleteByID(spec.Context(), release.Flag{}, flag.ID)) }()

			ff := &FixtureFactoryForPilots{
				FixtureFactory: spec.FixtureFactory,
				FlagID:         flag.ID,
			}

			specs.CommonSpec{
				EntityType:     release.Pilot{},
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

		t.Run(`ReleaseFlag`, func(t *testing.T) {
			specs.CommonSpec{
				EntityType:     release.Flag{},
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
					return spec.Subject.Save(spec.Context(), t.I(`flag`).(*release.Flag))
				}

				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.Truncate(spec.Context(), release.Flag{}))
				})

				s.Let(`flag`, func(t *testcase.T) interface{} {
					return &release.Flag{
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

			flag := spec.FixtureFactory.Create(release.Flag{}).(*release.Flag)
			require.Nil(t, spec.Subject.Save(spec.Context(), flag))
			defer func() { require.Nil(t, spec.Subject.Truncate(spec.Context(), release.Flag{})) }()

			ff := &FixtureFactoryForPilots{
				FixtureFactory: spec.FixtureFactory,
				FlagID:         flag.ID,
			}

			specs.CommonSpec{
				EntityType:     release.Pilot{},
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
	case release.Pilot:
		pilot := ff.FixtureFactory.Create(EntityType).(*release.Pilot)
		pilot.FlagID = ff.FlagID
		return pilot

	default:
		return ff.FixtureFactory.Create(EntityType)
	}
}
