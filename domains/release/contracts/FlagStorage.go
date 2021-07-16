package contracts

import (
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
	FixtureFactory func(testing.TB) contracts.FixtureFactory
}

func (c FlagStorage) storage() testcase.Var {
	return testcase.Var{
		Name: "release flag storage",
		Init: func(t *testcase.T) interface{} {
			return c.Subject(t).ReleaseFlag(sh.ContextGet(t))
		},
	}
}

func (c FlagStorage) storageGet(t *testcase.T) release.FlagStorage {
	return c.storage().Get(t).(release.FlagStorage)
}

func (c FlagStorage) String() string {
	return "FlagStorage"
}

func (c FlagStorage) Test(t *testing.T) {
	c.Spec(testcase.NewSpec(t))
}

func (c FlagStorage) Benchmark(b *testing.B) {
	c.Spec(testcase.NewSpec(b))
}

func (c FlagStorage) Spec(s *testcase.Spec) {
	sh.FixtureFactoryLet(s, c.FixtureFactory)

	newStorage := func(tb testing.TB) release.FlagStorage {
		return c.Subject(tb).ReleaseFlag(c.FixtureFactory(tb).Context())
	}

	T := release.Flag{}
	testcase.RunContract(s,
		contracts.Creator{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return newStorage(tb)
			},
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Finder{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return newStorage(tb)
			},
			FixtureFactory: c.FixtureFactory,
		},
		FlagFinder{
			Subject: func(tb testing.TB) release.FlagStorage {
				return newStorage(tb)
			},
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Updater{T: T,
			Subject: func(tb testing.TB) contracts.UpdaterSubject {
				return newStorage(tb)
			},
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Deleter{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return newStorage(tb)
			},
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Publisher{T: T,
			Subject: func(tb testing.TB) contracts.PublisherSubject {
				return newStorage(tb)
			},
			FixtureFactory: c.FixtureFactory,
		},
		contracts.OnePhaseCommitProtocol{T: release.Flag{},
			Subject: func(tb testing.TB) (frameless.OnePhaseCommitProtocol, contracts.CRD) {
				storage := c.Subject(tb)
				return storage, storage.ReleaseFlag(c.FixtureFactory(tb).Context())
			},
			FixtureFactory: c.FixtureFactory,
		},
	)

	s.Describe(`name release flag uniqueness across storage`, c.specFlagIsUniq)
}

func (c FlagStorage) specFlagIsUniq(s *testcase.Spec) {
	subject := func(t *testcase.T) error {
		return c.storageGet(t).Create(sh.FixtureFactoryGet(t).Context(), t.I(`flag`).(*release.Flag))
	}

	s.Before(func(t *testcase.T) {
		contracts.DeleteAllEntity(t, c.storageGet(t), sh.FixtureFactoryGet(t).Context())
	})

	flag := s.Let(`flag`, func(t *testcase.T) interface{} {
		return &release.Flag{Name: `my-uniq-flag-name`}
	})

	s.When(`flag already stored`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, subject(t))
			contracts.HasEntity(t, c.storageGet(t), sh.FixtureFactoryGet(t).Context(), flag.Get(t).(*release.Flag))
		})

		s.Then(`saving again will create error`, func(t *testcase.T) {
			require.Error(t, subject(t))
		})
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FlagFinder struct {
	Subject        func(testing.TB) release.FlagStorage
	FixtureFactory func(testing.TB) contracts.FixtureFactory
}

func (c FlagFinder) storage() testcase.Var {
	return testcase.Var{
		Name: "release FlagFinder storage",
		Init: func(t *testcase.T) interface{} {
			return c.Subject(t)
		},
	}
}

func (c FlagFinder) storageGet(t *testcase.T) release.FlagStorage {
	return c.storage().Get(t).(release.FlagStorage)
}

func (c FlagFinder) Test(t *testing.T) {
	c.Spec(testcase.NewSpec(t))
}

func (c FlagFinder) Benchmark(b *testing.B) {
	c.Spec(testcase.NewSpec(b))
}

func (c FlagFinder) cleanup(s *testcase.Spec) {
	once := &sync.Once{}
	s.Before(func(t *testcase.T) {
		once.Do(func() {
			contracts.DeleteAllEntity(t, c.storageGet(t), sh.FixtureFactoryGet(t).Context())
		})
	})
}

func (c FlagFinder) Spec(s *testcase.Spec) {
	sh.FixtureFactoryLet(s, c.FixtureFactory)
	c.cleanup(s)
	s.Describe(`.FindReleaseFlagByName`, c.specFindReleaseFlagByName)
	s.Describe(`.FindReleaseFlagsByName`, c.specFindReleaseFlagsByName)
}

func (c FlagFinder) specFindReleaseFlagsByName(s *testcase.Spec) {

	var (
		flagNames    = testcase.Var{Name: `flag names`}
		flagNamesGet = func(t *testcase.T) []string { return flagNames.Get(t).([]string) }
		subject      = func(t *testcase.T) iterators.Interface {
			flagEntriesIter := c.storageGet(t).FindByNames(sh.FixtureFactoryGet(t).Context(), flagNamesGet(t)...)
			t.Defer(flagEntriesIter.Close)
			return flagEntriesIter
		}
	)

	s.Before(func(t *testcase.T) {
		for _, name := range []string{`A`, `B`, `C`} {
			var flag = release.Flag{Name: name}
			contracts.CreateEntity(t, c.storageGet(t), sh.FixtureFactoryGet(t).Context(), &flag)
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

func (c FlagFinder) specFindReleaseFlagByName(s *testcase.Spec) {
	var (
		flagName    = s.LetValue(`flag name`, fixtures.Random.String())
		flagNameGet = func(t *testcase.T) string { return flagName.Get(t).(string) }
		subject     = func(t *testcase.T) *release.Flag {
			ff, err := c.storageGet(t).FindByName(sh.FixtureFactoryGet(t).Context(), flagNameGet(t))
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
			contracts.CreateEntity(t, c.storageGet(t), sh.FixtureFactoryGet(t).Context(), f)
			return f
		}).EagerLoading(s)

		s.Then(`searching for it returns the flag entity`, func(t *testcase.T) {
			ff := flag.Get(t).(*release.Flag)
			actually, err := c.storageGet(t).FindByName(sh.FixtureFactoryGet(t).Context(), ff.Name)
			require.Nil(t, err)
			require.Equal(t, ff, actually)
		})
	})
}
