package contracts

import (
	"context"
	"github.com/adamluzsi/frameless"
	"sync"
	"testing"

	"github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	sh "github.com/toggler-io/toggler/spechelper"
)

type FlagStorage struct {
	Subject        func(testing.TB) release.Storage
	FixtureFactory sh.FixtureFactory
}

func (spec FlagStorage) storage() testcase.Var {
	return testcase.Var{
		Name: "release flag storage",
		Init: func(t *testcase.T) interface{} {
			return spec.Subject(t).ReleaseFlag(sh.ContextGet(t))
		},
	}
}

func (spec FlagStorage) storageGet(t *testcase.T) release.FlagStorage {
	return spec.storage().Get(t).(release.FlagStorage)
}

func (spec FlagStorage) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec FlagStorage) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec FlagStorage) Spec(tb testing.TB) {
	testcase.NewSpec(tb).Describe(`FlagStorage`, func(s *testcase.Spec) {

		newStorage := func(tb testing.TB) release.FlagStorage {
			return spec.Subject(tb).ReleaseFlag(spec.FixtureFactory.Context())
		}

		T := release.Flag{}
		testcase.RunContract(s,
			contracts.Creator{T: T,
				Subject: func(tb testing.TB) contracts.CRD {
					return newStorage(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Finder{T: T,
				Subject: func(tb testing.TB) contracts.CRD {
					return newStorage(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			FlagFinder{
				Subject: func(tb testing.TB) release.FlagStorage {
					return newStorage(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Updater{T: T,
				Subject: func(tb testing.TB) contracts.UpdaterSubject {
					return newStorage(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Deleter{T: T,
				Subject: func(tb testing.TB) contracts.CRD {
					return newStorage(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Publisher{T: T,
				Subject: func(tb testing.TB) contracts.PublisherSubject {
					return newStorage(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.OnePhaseCommitProtocol{T: release.Flag{},
				Subject: func(tb testing.TB) (frameless.OnePhaseCommitProtocol, contracts.CRD) {
					storage := spec.Subject(tb)
					return storage, storage.ReleaseFlag(spec.FixtureFactory.Context())
				},
				FixtureFactory: spec.FixtureFactory,
			},
		)

		s.Describe(`name release flag uniqueness across storage`, spec.specFlagIsUniq)
	})
}

func (spec FlagStorage) specFlagIsUniq(s *testcase.Spec) {
	subject := func(t *testcase.T) error {
		return spec.storageGet(t).Create(spec.FixtureFactory.Context(), t.I(`flag`).(*release.Flag))
	}

	s.Before(func(t *testcase.T) {
		contracts.DeleteAllEntity(t, spec.storageGet(t), spec.FixtureFactory.Context())
	})

	flag := s.Let(`flag`, func(t *testcase.T) interface{} {
		return &release.Flag{Name: `my-uniq-flag-name`}
	})

	s.When(`flag already stored`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, subject(t))
			contracts.HasEntity(t, spec.storageGet(t), spec.FixtureFactory.Context(), flag.Get(t).(*release.Flag))
		})

		s.Then(`saving again will create error`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FlagFinder struct {
	Subject func(testing.TB) release.FlagStorage

	contracts.FixtureFactory
}

func (spec FlagFinder) storage() testcase.Var {
	return testcase.Var{
		Name: "release FlagFinder storage",
		Init: func(t *testcase.T) interface{} {
			return spec.Subject(t)
		},
	}
}

func (spec FlagFinder) storageGet(t *testcase.T) release.FlagStorage {
	return spec.storage().Get(t).(release.FlagStorage)
}

func (spec FlagFinder) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec FlagFinder) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec FlagFinder) cleanup(s *testcase.Spec) {
	once := &sync.Once{}
	s.Before(func(t *testcase.T) {
		once.Do(func() {
			contracts.DeleteAllEntity(t, spec.storageGet(t), spec.Context())
		})
	})
}

func (spec FlagFinder) Spec(tb testing.TB) {
	testcase.NewSpec(tb).Describe(`FlagFinder`, func(s *testcase.Spec) {
		spec.cleanup(s)
		s.Describe(`.FindReleaseFlagByName`, spec.specFindReleaseFlagByName)
		s.Describe(`.FindReleaseFlagsByName`, spec.specFindReleaseFlagsByName)
	})
}

func (spec FlagFinder) specFindReleaseFlagsByName(s *testcase.Spec) {

	var (
		flagNames    = testcase.Var{Name: `flag names`}
		flagNamesGet = func(t *testcase.T) []string { return flagNames.Get(t).([]string) }
		subject      = func(t *testcase.T) iterators.Interface {
			flagEntriesIter := spec.storageGet(t).FindReleaseFlagsByName(spec.context(), flagNamesGet(t)...)
			t.Defer(flagEntriesIter.Close)
			return flagEntriesIter
		}
	)

	s.Before(func(t *testcase.T) {
		for _, name := range []string{`A`, `B`, `C`} {
			var flag = release.Flag{Name: name}
			contracts.CreateEntity(t, spec.storageGet(t), spec.Context(), &flag)
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
		flagNames.Let(s, func(t *testcase.T) interface{} {
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
		flagNames.Let(s, func(t *testcase.T) interface{} {
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
		flagNames.Let(s, func(t *testcase.T) interface{} {
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
	var (
		flagName    = s.LetValue(`flag name`, fixtures.Random.String())
		flagNameGet = func(t *testcase.T) string { return flagName.Get(t).(string) }
		subject     = func(t *testcase.T) *release.Flag {
			ff, err := spec.storageGet(t).FindReleaseFlagByName(spec.context(), flagNameGet(t))
			require.Nil(t, err)
			return ff
		}
	)

	s.Then(`we receive back nil pointer`, func(t *testcase.T) {
		require.Nil(t, subject(t))
	})

	s.When(`we have a release flag already set`, func(s *testcase.Spec) {
		flag := s.Let(`flag`, func(t *testcase.T) interface{} {
			f := &release.Flag{Name: flagNameGet(t)}
			contracts.CreateEntity(t, spec.storageGet(t), spec.Context(), f)
			return f
		}).EagerLoading(s)

		s.Then(`searching for it returns the flag entity`, func(t *testcase.T) {
			ff := flag.Get(t).(*release.Flag)
			actually, err := spec.storageGet(t).FindReleaseFlagByName(spec.context(), ff.Name)
			require.Nil(t, err)
			require.Equal(t, ff, actually)
		})
	})
}

func (spec FlagFinder) context() context.Context {
	return spec.FixtureFactory.Context()
}
