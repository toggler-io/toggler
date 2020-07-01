package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	. "github.com/toggler-io/toggler/testing"
)

type StorageSpec struct {
	Subject release.Storage
	specs.FixtureFactory
}

func (spec StorageSpec) Test(t *testing.T) {
	t.Run(`releases`, func(t *testing.T) {
		RolloutStorageSpec{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		}.Test(t)

		t.Run(`Flag`, func(t *testing.T) {
			specs.CommonSpec{
				EntityType:     release.Flag{},
				FixtureFactory: spec.FixtureFactory,
				Subject:        spec.Subject,
			}.Test(t)

			specs.OnePhaseCommitProtocolSpec{
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
					return spec.Subject.Create(spec.Context(), t.I(`flag`).(*release.Flag))
				}

				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.DeleteAll(spec.Context(), release.Flag{}))
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

		t.Run(`ManualPilot`, func(t *testing.T) {
			s := testcase.NewSpec(t)
			SetUp(s)

			s.Let(ExampleStorageLetVar, func(t *testcase.T) interface{} {
				return spec.Subject
			})

			s.After(func(t *testcase.T) {
				require.Nil(t, spec.Subject.DeleteAll(spec.Context(), release.Flag{}))
			})

			pilotFinderSpec{
				FixtureFactory: spec.FixtureFactory,
				Subject:        spec.Subject,
			}.Test(t)
		})
	})
}

func (spec StorageSpec) Benchmark(b *testing.B) {
	b.Run(`releases`, func(b *testing.B) {
		b.Skip()
	})
}
