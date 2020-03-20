package specs

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
)

type StorageSpec struct {
	Subject release.Storage
	specs.FixtureFactory
}

func (spec StorageSpec) Benchmark(b *testing.B) {
	b.Run(`releases`, func(b *testing.B) {

		b.Run(`Flag`, func(b *testing.B) {
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
			defer func() { require.Nil(b, spec.Subject.Truncate(spec.Context(), release.Flag{})) }()
			ff := &FixtureFactoryForPilots{
				FixtureFactory: spec.FixtureFactory,
				GetFlagID: func() string {
					f := spec.FixtureFactory.Create(release.Flag{}).(*release.Flag)
					require.Nil(b, spec.Subject.Create(spec.Context(), f))
					return f.ID
				},
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

		b.Run(`IPAllow`, func(b *testing.B) {
			b.Skip()
		})

	})
}

func (spec StorageSpec) Test(t *testing.T) {
	t.Run(`releases`, func(t *testing.T) {
		t.Run(`Flag`, func(t *testing.T) {
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
					return spec.Subject.Create(spec.Context(), t.I(`flag`).(*release.Flag))
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
			defer func() { require.Nil(t, spec.Subject.Truncate(spec.Context(), release.Flag{})) }()

			ff := &FixtureFactoryForPilots{
				FixtureFactory: spec.FixtureFactory,
				GetFlagID: func() string {
					f := spec.FixtureFactory.Create(release.Flag{}).(*release.Flag)
					require.Nil(t, spec.Subject.Create(spec.Context(), f))
					return f.ID
				},
			}

			specs.CommonSpec{
				EntityType:     release.Pilot{},
				FixtureFactory: ff,
				Subject:        spec.Subject,
			}.Test(t)

			pilotFinderSpec{
				FixtureFactory: ff,
				Subject:        spec.Subject,
			}.Test(t)

		})

		t.Run(`IPAllow`, func(t *testing.T) {
			defer func() { require.Nil(t, spec.Subject.Truncate(spec.Context(), release.Flag{})) }()

			ff := &FixtureFactoryForAllows{
				FixtureFactory: spec.FixtureFactory,
				GetFlagID: func() string {
					flag := spec.FixtureFactory.Create(release.Flag{}).(*release.Flag)
					require.Nil(t, spec.Subject.Create(spec.Context(), flag))
					return flag.ID
				},
			}

			specs.CommonSpec{
				EntityType:     release.IPAllow{},
				FixtureFactory: ff,
				Subject:        spec.Subject,
			}.Test(t)

			AllowFinderSpec{
				Subject:        spec.Subject,
				FixtureFactory: ff,
			}.Test(t)

			t.Run(`multiple allow entry can be stored for one flag`, func(t *testing.T) {
				require.Nil(t, spec.Subject.Truncate(spec.Context(), release.IPAllow{}))
				require.Nil(t, spec.Subject.Truncate(spec.Context(), release.Flag{}))
				defer func() {
					require.Nil(t, spec.Subject.Truncate(spec.Context(), release.Flag{}))
					require.Nil(t, spec.Subject.Truncate(spec.Context(), release.IPAllow{}))
				}()

				flagID := ff.GetFlagID()

				a1 := ff.Create(release.IPAllow{}).(*release.IPAllow)
				a1.FlagID = flagID
				a2 := ff.Create(release.IPAllow{}).(*release.IPAllow)
				a2.FlagID = flagID

				require.Nil(t, spec.Subject.Create(spec.Context(), a1))
				require.Nil(t, spec.Subject.Create(spec.Context(), a2))
			})
		})
	})
}

type FixtureFactoryForPilots struct {
	specs.FixtureFactory
	GetFlagID func() string
}

func (ff *FixtureFactoryForPilots) Create(EntityType interface{}) interface{} {
	switch reflects.BaseValueOf(EntityType).Interface().(type) {
	case release.Pilot:
		pilot := ff.FixtureFactory.Create(EntityType).(*release.Pilot)
		pilot.FlagID = ff.GetFlagID()
		return pilot

	default:
		return ff.FixtureFactory.Create(EntityType)
	}
}

type FixtureFactoryForAllows struct {
	specs.FixtureFactory
	GetFlagID func() string
}

func (ff *FixtureFactoryForAllows) Create(EntityType interface{}) interface{} {
	switch reflects.BaseValueOf(EntityType).Interface().(type) {
	case release.IPAllow:
		allow := ff.FixtureFactory.Create(EntityType).(*release.IPAllow)
		allow.FlagID = ff.GetFlagID()
		allow.InternetProtocolAddress = fmt.Sprintf(`192.168.1.%d`, rand.Intn(255))
		return allow

	default:
		return ff.FixtureFactory.Create(EntityType)
	}
}
