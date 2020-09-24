package specs

import (
	"context"
	"testing"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	. "github.com/toggler-io/toggler/testing"
)

type FlagStorage struct {
	Subject interface {
		resources.Creator
		resources.Finder
		resources.Updater
		resources.Deleter
		resources.OnePhaseCommitProtocol
		resources.CreatorPublisher
		resources.UpdaterPublisher
		resources.DeleterPublisher
		release.FlagFinder
	}

	FixtureFactory FixtureFactory
}

func (spec FlagStorage) Test(t *testing.T) {
	t.Run(`FlagStorage`, func(t *testing.T) {
		spec.Spec(t)
	})
}

func (spec FlagStorage) Benchmark(b *testing.B) {
	b.Run(`FlagStorage`, func(b *testing.B) {
		spec.Spec(b)
	})
}

func (spec FlagStorage) Spec(tb testing.TB) {
	specs.Run(tb,
		specs.CRUD{
			EntityType:     release.Flag{},
			FixtureFactory: spec.FixtureFactory,
			Subject:        spec.Subject,
		},
		FlagFinder{
			Subject:        spec.Subject,
			FixtureFactory: spec.FixtureFactory,
		},
		specs.OnePhaseCommitProtocol{
			EntityType:     release.Flag{},
			FixtureFactory: spec.FixtureFactory,
			Subject:        spec.Subject,
		},
		specs.CreatorPublisher{
			Subject:        spec.Subject,
			EntityType:     release.Flag{},
			FixtureFactory: spec.FixtureFactory,
		},
		specs.UpdaterPublisher{
			Subject:        spec.Subject,
			EntityType:     release.Flag{},
			FixtureFactory: spec.FixtureFactory,
		},
		specs.DeleterPublisher{
			Subject:        spec.Subject,
			EntityType:     release.Flag{},
			FixtureFactory: spec.FixtureFactory,
		},
	)

	s := testcase.NewSpec(tb)
	s.Describe(`name release flag uniqueness across storage`, spec.specFlagIsUniq)
}

func (spec FlagStorage) specFlagIsUniq(s *testcase.Spec) {
	subject := func(t *testcase.T) error {
		return spec.Subject.Create(spec.FixtureFactory.Context(), t.I(`flag`).(*release.Flag))
	}

	s.Before(func(t *testcase.T) {
		require.Nil(t, spec.Subject.DeleteAll(spec.FixtureFactory.Context(), release.Flag{}))
	})

	s.Let(`flag`, func(t *testcase.T) interface{} {
		return &release.Flag{Name: `my-uniq-flag-name`}
	})

	s.When(`flag already stored`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { require.Nil(t, subject(t)) })

		s.Then(`saving again will create error`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FlagFinder struct {
	Subject interface {
		release.FlagFinder
		resources.Creator
		resources.Finder
		resources.Deleter
	}

	specs.FixtureFactory
}

func (spec FlagFinder) Test(t *testing.T) {
	t.Run(`FlagFinder`, func(t *testing.T) {
		spec.Spec(t)
	})
}

func (spec FlagFinder) Benchmark(b *testing.B) {
	b.Run(`FlagFinder`, func(b *testing.B) {
		spec.Spec(b)
	})
}

func (spec FlagFinder) setup(s *testcase.Spec) {
	s.LetValue(`flag name`, fixtures.Random.String())

	s.Around(func(t *testcase.T) func() {
		require.Nil(t, spec.Subject.DeleteAll(spec.context(), release.Flag{}))
		return func() { require.Nil(t, spec.Subject.DeleteAll(spec.context(), release.Flag{})) }
	})
}

func (spec FlagFinder) flagName(t *testcase.T) string {
	return t.I(`flag name`).(string)
}

func (spec FlagFinder) Spec(tb testing.TB) {
	s := testcase.NewSpec(tb)
	spec.setup(s)
	s.Describe(`FindReleaseFlagByName`, spec.specFindReleaseFlagByName)
	s.Describe(`FindReleaseFlagsByName`, spec.specFindReleaseFlagsByName)
}

func (spec FlagFinder) specFindReleaseFlagsByName(s *testcase.Spec) {
	var subject = func(t *testcase.T) iterators.Interface {
		flagEntriesIter := spec.Subject.FindReleaseFlagsByName(spec.context(), t.I(`flag names`).([]string)...)
		t.Defer(flagEntriesIter.Close)
		return flagEntriesIter
	}

	s.Before(func(t *testcase.T) {
		ctx := spec.context()
		for _, name := range []string{`A`, `B`, `C`} {
			var flag = release.Flag{Name: name}
			require.Nil(t, spec.Subject.Create(ctx, &flag))
			t.Defer(spec.Subject.DeleteByID, ctx, flag, flag.ID)
		}
	})

	mustContainName := func(t *testcase.T, ffs []release.Flag, name string) {
		for _, ff := range ffs {
			if ff.Name == name {
				return
			}
		}

		t.Fatalf(`flag name could not be found: %s`, name)
	}

	s.When(`we request flags that can be found`, func(s *testcase.Spec) {
		s.Let(`flag names`, func(t *testcase.T) interface{} {
			return []string{`A`, `B`, `C`}
		})

		s.Then(`it will return all of them`, func(t *testcase.T) {
			flagsIter := subject(t)

			var flags []release.Flag
			require.Nil(t, iterators.Collect(flagsIter, &flags))

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

		s.Then(`it will return existing flags`, func(t *testcase.T) {
			flagsIter := subject(t)

			var flags []release.Flag
			require.Nil(t, iterators.Collect(flagsIter, &flags))

			t.Logf(`%#v`, flags)

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
}

func (spec FlagFinder) specFindReleaseFlagByName(s *testcase.Spec) {
	subject := func(t *testcase.T) *release.Flag {
		ff, err := spec.Subject.FindReleaseFlagByName(spec.context(), spec.flagName(t))
		require.Nil(t, err)
		return ff
	}

	s.Before(func(t *testcase.T) {
		require.Nil(t, spec.Subject.DeleteAll(spec.context(), release.Flag{}))
	})

	s.Then(`we receive back nil pointer`, func(t *testcase.T) {
		require.Nil(t, subject(t))
	})

	s.When(`we have a release flag already set`, func(s *testcase.Spec) {
		s.Let(`ff`, func(t *testcase.T) interface{} {
			flag := &release.Flag{Name: spec.flagName(t)}
			require.Nil(t, spec.Subject.Create(spec.context(), flag))
			t.Defer(spec.Subject.DeleteByID, spec.context(), *flag, flag.ID)
			return flag
		})

		s.Then(`searching for it returns the flag entity`, func(t *testcase.T) {
			ff := t.I(`ff`).(*release.Flag)
			actually, err := spec.Subject.FindReleaseFlagByName(spec.context(), ff.Name)
			require.Nil(t, err)
			require.Equal(t, ff, actually)
		})
	})
}

func (spec FlagFinder) context() context.Context {
	return spec.FixtureFactory.Context()
}
