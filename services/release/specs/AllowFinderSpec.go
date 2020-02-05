package specs

import (
	"testing"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/services/release"
)

type AllowFinderSpec struct {
	Subject interface {
		release.AllowFinder
		specs.MinimumRequirements
	}

	specs.FixtureFactory
}

func (spec AllowFinderSpec) Benchmark(b *testing.B) {
	b.Run(`AllowFinderSpec`, func(b *testing.B) {
		b.Skip(`TODO`)
	})
}

func (spec AllowFinderSpec) Test(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Describe(`AllowFinderSpec`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, spec.Subject.Truncate(spec.Context(), release.Flag{}))
			require.Nil(t, spec.Subject.Truncate(spec.Context(), release.IPAllow{}))
		})

		s.Describe(`FindReleaseAllowsByReleaseFlags`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) release.AllowEntries {
				return spec.Subject.FindReleaseAllowsByReleaseFlags(spec.Context(), *t.I(`flag`).(*release.Flag))
			}

			s.Let(`flag`, func(t *testcase.T) interface{} {
				return spec.Create(release.Flag{})
			})
			s.Around(func(t *testcase.T) func() {
				f := t.I(`flag`).(*release.Flag)
				require.Nil(t, spec.Subject.Create(spec.Context(), f))
				return func() { require.Nil(t, spec.Subject.DeleteByID(spec.Context(), release.Flag{}, f.ID)) }
			})

			s.When(`no allow stored in the resource`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.Truncate(spec.Context(), release.IPAllow{}))
				})

				s.Then(`it will result in an empty list`, func(t *testcase.T) {
					count, err := iterators.Count(subject(t))
					require.Nil(t, err)
					require.Equal(t, 0, count)
				})
			})

			s.When(`allow is saved in the storage`, func(s *testcase.Spec) {
				s.Let(`allow`, func(t *testcase.T) interface{} {
					return spec.Create(release.IPAllow{}).(*release.IPAllow)
				})
				s.Around(func(t *testcase.T) func() {
					var tds []func()

					f := t.I(`allow's flag`).(*release.Flag)
					if f.ID == `` { // don't try to save again if flag is already saved
						require.Nil(t, spec.Subject.Create(spec.Context(), f))
						tds = append(tds, func() {
							require.Nil(t, spec.Subject.DeleteByID(spec.Context(), release.Flag{}, f.ID))
						})
					}

					a := t.I(`allow`).(*release.IPAllow)
					a.FlagID = f.ID
					require.Nil(t, spec.Subject.Create(spec.Context(), a))
					tds = append(tds, func() {
						require.Nil(t, spec.Subject.DeleteByID(spec.Context(), release.IPAllow{}, a.ID))
					})

					return func() {
						for _, td := range tds {
							td()
						}
					}
				})

				s.And(`the allow belongs to the flag the method received`, func(s *testcase.Spec) {
					s.Let(`allow's flag`, func(t *testcase.T) interface{} { return t.I(`flag`) })

					s.Then(`it will return the allow flag`, func(t *testcase.T) {
						i := subject(t)

						expected := t.I(`allow`).(*release.IPAllow)
						var actual release.IPAllow

						require.True(t, i.Next())
						require.Nil(t, i.Decode(&actual))
						require.False(t, i.Next())
						require.Nil(t, i.Err())
						require.Equal(t, expected, &actual)
					})
					
					s.And(`the allow ip address is empty`, func(s *testcase.Spec) {
						s.Let(`allow`, func(t *testcase.T) interface{} {
							a := spec.Create(release.IPAllow{}).(*release.IPAllow)
							a.InternetProtocolAddress = ``
							return a
						})

						s.Then(`it will return the allow flag with empty ip addr`, func(t *testcase.T) {
							i := subject(t)

							expected := t.I(`allow`).(*release.IPAllow)
							var actual release.IPAllow

							require.True(t, i.Next())
							require.Nil(t, i.Decode(&actual))
							require.False(t, i.Next())
							require.Nil(t, i.Err())
							require.Equal(t, expected, &actual)
						})
					})
				})

				s.And(`the allow belongs to a different flag from the one the method received`, func(s *testcase.Spec) {
					s.Let(`allow's flag`, func(t *testcase.T) interface{} {
						return spec.Create(release.Flag{})
					})

					s.Then(`it will not return the release-allows that belong to other release-flag`, func(t *testcase.T) {
						count, err := iterators.Count(subject(t))
						require.Nil(t, err)
						require.Equal(t, 0, count)
					})
				})

			})
		})
	})
}
