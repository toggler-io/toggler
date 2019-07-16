package specs

import (
	"context"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/toggler/testing"
	"testing"

	"github.com/adamluzsi/toggler/services/rollouts"

	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/stretchr/testify/require"
)

type FlagFinderSpec struct {
	Subject interface {
		rollouts.FlagFinder

		specs.MinimumRequirements
	}
}

func (spec FlagFinderSpec) Test(t *testing.T) {
	s := testcase.NewSpec(t)

	featureName := ExampleFeatureName()

	s.Describe(`FlagFinderSpec`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, spec.Subject.Truncate(context.Background(), rollouts.FeatureFlag{}))
		})

		s.After(func(t *testcase.T) {
			require.Nil(t, spec.Subject.Truncate(context.Background(), rollouts.FeatureFlag{}))
		})

		s.Describe(`FindFlagByName`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) *rollouts.FeatureFlag {
				ff, err := spec.Subject.FindFlagByName(context.Background(), featureName)
				require.Nil(t, err)
				return ff
			}

			s.When(`we don't have feature flag yet`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) { require.Nil(t, spec.Subject.Truncate(context.Background(), rollouts.FeatureFlag{})) })

				s.Then(`we receive back nil pointer`, func(t *testcase.T) {
					require.Nil(t, subject(t))
				})
			})

			s.When(`we have a feature flag already set`, func(s *testcase.Spec) {
				s.Let(`ff`, func(t *testcase.T) interface{} {
					return &rollouts.FeatureFlag{Name: featureName}
				})

				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.Save(context.Background(), t.I(`ff`).(*rollouts.FeatureFlag)))
				})

				s.Then(`searching for it returns the flag entity`, func(t *testcase.T) {
					ff := t.I(`ff`).(*rollouts.FeatureFlag)
					actually, err := spec.Subject.FindFlagByName(context.Background(), ff.Name)
					require.Nil(t, err)
					require.Equal(t, ff, actually)
				})
			})
		})

		s.Describe(`FindFlagsByName`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) frameless.Iterator {
				return spec.Subject.FindFlagsByName(context.Background(), t.I(`flag names`).([]string)...)
			}

			s.Before(func(t *testcase.T) {
				ctx := context.Background()
				require.Nil(t, spec.Subject.Save(ctx, &rollouts.FeatureFlag{Name: `A`}))
				require.Nil(t, spec.Subject.Save(ctx, &rollouts.FeatureFlag{Name: `B`}))
				require.Nil(t, spec.Subject.Save(ctx, &rollouts.FeatureFlag{Name: `C`}))
			})

			mustContainName := func(t *testcase.T, ffs []rollouts.FeatureFlag, name string) {
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

					var flags []rollouts.FeatureFlag
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

					var flags []rollouts.FeatureFlag
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
