package specs

import (
	"context"
	"testing"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/services/release"
	. "github.com/toggler-io/toggler/testing"
	"github.com/stretchr/testify/require"
)

type FlagFinderSpec struct {
	Subject interface {
		release.FlagFinder

		specs.MinimumRequirements
	}

	specs.FixtureFactory
}

func (spec FlagFinderSpec) Benchmark(b *testing.B) {
	b.Run(`FlagFinderSpec`, func(b *testing.B) {
		b.Skip(`TODO`)
	})
}

func (spec FlagFinderSpec) Test(t *testing.T) {
	s := testcase.NewSpec(t)

	featureName := ExampleName()

	s.Describe(`FlagFinderSpec`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, spec.Subject.Truncate(spec.ctx(), release.Flag{}))
		})

		s.After(func(t *testcase.T) {
			require.Nil(t, spec.Subject.Truncate(spec.ctx(), release.Flag{}))
		})

		s.Describe(`FindReleaseFlagByName`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) *release.Flag {
				ff, err := spec.Subject.FindReleaseFlagByName(spec.ctx(), featureName)
				require.Nil(t, err)
				return ff
			}

			s.When(`we don't have feature flag yet`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) { require.Nil(t, spec.Subject.Truncate(spec.ctx(), release.Flag{})) })

				s.Then(`we receive back nil pointer`, func(t *testcase.T) {
					require.Nil(t, subject(t))
				})
			})

			s.When(`we have a feature flag already set`, func(s *testcase.Spec) {
				s.Let(`ff`, func(t *testcase.T) interface{} {
					return &release.Flag{Name: featureName}
				})

				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.Save(spec.ctx(), t.I(`ff`).(*release.Flag)))
				})

				s.Then(`searching for it returns the flag entity`, func(t *testcase.T) {
					ff := t.I(`ff`).(*release.Flag)
					actually, err := spec.Subject.FindReleaseFlagByName(spec.ctx(), ff.Name)
					require.Nil(t, err)
					require.Equal(t, ff, actually)
				})
			})
		})

		s.Describe(`FindFlagsByName`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) frameless.Iterator {
				return spec.Subject.FindFlagsByName(spec.ctx(), t.I(`flag names`).([]string)...)
			}

			s.Before(func(t *testcase.T) {
				ctx := spec.ctx()
				require.Nil(t, spec.Subject.Save(ctx, &release.Flag{Name: `A`}))
				require.Nil(t, spec.Subject.Save(ctx, &release.Flag{Name: `B`}))
				require.Nil(t, spec.Subject.Save(ctx, &release.Flag{Name: `C`}))
			})

			mustContainName := func(t *testcase.T, ffs []release.Flag, name string) {
				for _, ff := range ffs {
					if ff.Name == name {
						return
					}
				}

				t.Fatalf(`featoure flag name could not be found: %s`, name)
			}

			s.When(`we request flags that can be found`, func(s *testcase.Spec) {
				s.Let(`flag names`, func(t *testcase.T) interface{} {
					return []string{`A`, `B`, `C`}
				})

				s.Then(`it will return all of them`, func(t *testcase.T) {
					flagsIter := subject(t)

					var flags []release.Flag
					require.Nil(t, iterators.CollectAll(flagsIter, &flags))

					require.Equal(t, 3, len(flags))
					mustContainName(t, flags, `A`)
					mustContainName(t, flags, `B`)
					mustContainName(t, flags, `C`)
				})
			})

			s.When(`the requested flags only partially found`, func(s *testcase.Spec) {
				s.Let(`flag names`, func(t *testcase.T) interface{} {
					return []string{`A`, `B`, `D`}
				})

				s.Then(`it will return all of them`, func(t *testcase.T) {
					flagsIter := subject(t)

					var flags []release.Flag
					require.Nil(t, iterators.CollectAll(flagsIter, &flags))

					require.Equal(t, 2, len(flags))
					mustContainName(t, flags, `A`)
					mustContainName(t, flags, `B`)
				})
			})

			s.When(`none of the requested flags found`, func(s *testcase.Spec) {
				s.Let(`flag names`, func(t *testcase.T) interface{} {
					return []string{`R`, `O`, `F`, `L`}
				})

				s.Then(`it will return an empty iterator`, func(t *testcase.T) {
					flagsIter := subject(t)

					count, err := iterators.Count(flagsIter)
					require.Nil(t, err)
					require.Equal(t, 0, count)
				})
			})

		})
	})
}

func (spec FlagFinderSpec) ctx() context.Context {
	return spec.FixtureFactory.Context()
}
