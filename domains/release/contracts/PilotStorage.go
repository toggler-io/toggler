package contracts

import (
	"context"
	"github.com/adamluzsi/frameless"
	"math/rand"
	"strconv"
	"testing"

	"github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	sh "github.com/toggler-io/toggler/spechelper"
)

type PilotStorage struct {
	Subject        func(testing.TB) release.Storage
	FixtureFactory func(testing.TB) contracts.FixtureFactory
}

func (c PilotStorage) storage() testcase.Var {
	return testcase.Var{
		Name: "release pilot storage",
		Init: func(t *testcase.T) interface{} {
			return sh.StorageGet(t).ReleasePilot(sh.ContextGet(t))
		},
	}
}

func (c PilotStorage) storageGet(t *testcase.T) release.PilotStorage {
	return c.storage().Get(t).(release.PilotStorage)
}

func (c PilotStorage) context(t *testcase.T) context.Context {
	return sh.FixtureFactoryGet(t).Context()
}

func (c PilotStorage) String() string {
	return "PilotStorage"
}

func (c PilotStorage) Test(t *testing.T) {
	c.Spec(testcase.NewSpec(t))
}

func (c PilotStorage) Benchmark(b *testing.B) {
	c.Spec(testcase.NewSpec(b))
}

func (c PilotStorage) Spec(s *testcase.Spec) {
	sh.SetUp(s)
	sh.FixtureFactoryLet(s, c.FixtureFactory)

	// required for FixtureFactory.Dynamic
	sh.Storage.Let(s, func(t *testcase.T) interface{} {
		return c.Subject(t)
	})

	releasePilotStorage := func(tb testing.TB) release.PilotStorage {
		return c.Subject(tb).ReleasePilot(c.FixtureFactory(tb).Context())
	}

	T := release.Pilot{}
	testcase.RunContract(s,
		contracts.Creator{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return releasePilotStorage(tb)
			},
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Finder{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return releasePilotStorage(tb)
			},
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Updater{T: T,
			Subject: func(tb testing.TB) contracts.UpdaterSubject {
				return releasePilotStorage(tb)
			},
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Deleter{T: T,
			Subject: func(tb testing.TB) contracts.CRD {
				return releasePilotStorage(tb)
			},
			FixtureFactory: c.FixtureFactory,
		},
		contracts.Publisher{T: T,
			Subject: func(tb testing.TB) contracts.PublisherSubject {
				return releasePilotStorage(tb)
			},
			FixtureFactory: c.FixtureFactory,
		},
		contracts.OnePhaseCommitProtocol{T: release.Pilot{},
			Subject: func(tb testing.TB) (frameless.OnePhaseCommitProtocol, contracts.CRD) {
				storage := c.Subject(tb)
				return storage, storage.ReleasePilot(c.FixtureFactory(tb).Context())
			},
			FixtureFactory: c.FixtureFactory,
		},
	)

	s.Describe(`custom Find queries`, c.specPilotFinder)
}

func (c PilotStorage) specPilotFinder(s *testcase.Spec) {
	s.Describe(`ManualPilotFinder`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			contracts.DeleteAllEntity(t, c.storageGet(t), c.context(t))
		})
		s.After(func(t *testcase.T) {
			contracts.DeleteAllEntity(t, c.storageGet(t), c.context(t))
		})

		s.Describe(`.FindByFlag`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) iterators.Interface {
				pilotEntriesIter := c.storageGet(t).FindByFlag(c.context(t), *sh.ExampleReleaseFlag(t))
				t.Defer(pilotEntriesIter.Close)
				return pilotEntriesIter
			}

			thenNoPilotsFound := func(s *testcase.Spec) {
				s.Then(`no pilots found`, func(t *testcase.T) {
					iter := subject(t)
					require.NotNil(t, iter)
					require.False(t, iter.Next())
					require.Nil(t, iter.Err())
					require.Nil(t, iter.Close())
				})
			}

			s.When(`flag was never persisted before`, func(s *testcase.Spec) {
				s.Let(sh.LetVarExampleReleaseFlag, func(t *testcase.T) interface{} {
					rf := sh.FixtureFactoryGet(t).Create(release.Flag{}).(release.Flag)
					return &rf
				})

				thenNoPilotsFound(s)
			})

			s.When(`flag is persisted`, func(s *testcase.Spec) {
				thenNoPilotsFound(s)

				s.And(`there are manual pilot configs for the release flag`, func(s *testcase.Spec) {
					expectedPilots := s.Let(`expectedPilots`, func(t *testcase.T) interface{} {
						var expectedPilots []*release.Pilot
						for i := 0; i < 5; i++ {
							pilot := &release.Pilot{
								FlagID:        sh.ExampleReleaseFlag(t).ID,
								EnvironmentID: sh.ExampleDeploymentEnvironment(t).ID,
								PublicID:      strconv.Itoa(i),
							}

							contracts.CreateEntity(t, c.storageGet(t), c.context(t), pilot)
							expectedPilots = append(expectedPilots, pilot)
						}
						return expectedPilots
					}).EagerLoading(s)

					expectedPilotsGet := func(t *testcase.T) []*release.Pilot {
						return expectedPilots.Get(t).([]*release.Pilot)
					}

					s.Then(`it will return all of them`, func(t *testcase.T) {
						iter := subject(t)
						defer iter.Close()
						require.NotNil(t, iter)
						var actualPilots []*release.Pilot
						for iter.Next() {
							var actually release.Pilot
							require.Nil(t, iter.Decode(&actually))
							actualPilots = append(actualPilots, &actually)
						}
						require.Nil(t, iter.Err())

						expectedPilots := expectedPilotsGet(t)
						require.True(t, len(expectedPilots) == len(actualPilots))
						require.ElementsMatch(t, expectedPilots, actualPilots)
					})
				})
			})
		})

		s.Describe(`FindReleaseManualPilotByExternalID`, func(s *testcase.Spec) {
			var subject = func(t *testcase.T) (*release.Pilot, error) {
				return c.storageGet(t).FindByFlagEnvPublicID(
					c.context(t),
					sh.ExampleReleaseFlag(t).ID,
					sh.ExampleDeploymentEnvironment(t).ID,
					sh.ExampleIDGet(t),
				)
			}

			s.Before(func(t *testcase.T) {
				contracts.DeleteAllEntity(t, c.storageGet(t), c.context(t))
			})

			ThenNoPilotsFound := func(s *testcase.Spec) {
				s.Then(`no pilots found`, func(t *testcase.T) {
					pilot, err := subject(t)
					require.Nil(t, err)
					require.Nil(t, pilot)
				})
			}

			s.When(`flag is not persisted`, func(s *testcase.Spec) {
				s.Let(sh.LetVarExampleReleaseFlag, func(t *testcase.T) interface{} {
					rf := sh.FixtureFactoryGet(t).Create(release.Flag{}).(release.Flag)
					return &rf
				})

				ThenNoPilotsFound(s)
			})

			s.When(`flag persisted already exists`, func(s *testcase.Spec) {
				s.Let(`featureFlagID`, func(t *testcase.T) interface{} {
					return sh.ExampleReleaseFlag(t).ID
				})

				ThenNoPilotsFound(s)

				s.And(`the given there is a registered pilot for the feature`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						pilot := &release.Pilot{
							FlagID:        sh.ExampleReleaseFlag(t).ID,
							EnvironmentID: sh.ExampleDeploymentEnvironment(t).ID,
							PublicID:      sh.ExampleIDGet(t),
						}
						contracts.CreateEntity(t, c.storageGet(t), c.context(t), pilot)
					})

					s.Then(`then pilots will be retrieved`, func(t *testcase.T) {
						pilot, err := subject(t)
						require.Nil(t, err)
						require.NotNil(t, pilot)

						require.Equal(t, sh.ExampleIDGet(t), pilot.PublicID)
						require.Equal(t, sh.ExampleReleaseFlag(t).ID, pilot.FlagID)
						require.Equal(t, sh.ExampleDeploymentEnvironment(t).ID, pilot.EnvironmentID)
					})
				})
			})
		})

		s.Describe(`FindReleasePilotsByExternalID`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) iterators.Interface {
				pilotEntriesIter := c.storageGet(t).FindByPublicID(c.context(t), sh.ExampleExternalPilotID(t))
				t.Defer(pilotEntriesIter.Close)
				return pilotEntriesIter
			}

			s.Let(`PilotExternalID`, func(t *testcase.T) interface{} {
				return fixtures.Random.String()
			})

			s.When(`there is no pilot records`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					contracts.DeleteAllEntity(t, c.storageGet(t), c.context(t))
				})

				s.Then(`it will return an empty result set`, func(t *testcase.T) {
					count, err := iterators.Count(subject(t))
					require.Nil(t, err)
					require.Equal(t, 0, count)
				})
			})

			s.When(`the given pilot id has no records`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					extID := fixtures.Random.String()
					var newUUID = func() string {
						uuidV4, err := uuid.NewRandom()
						require.Nil(t, err)
						return uuidV4.String()
					}

					contracts.CreateEntity(t, c.storageGet(t), c.context(t), &release.Pilot{FlagID: newUUID(), EnvironmentID: sh.ExampleDeploymentEnvironment(t).ID, PublicID: extID, IsParticipating: true})
					contracts.CreateEntity(t, c.storageGet(t), c.context(t), &release.Pilot{FlagID: newUUID(), EnvironmentID: sh.ExampleDeploymentEnvironment(t).ID, PublicID: extID, IsParticipating: true})
					contracts.CreateEntity(t, c.storageGet(t), c.context(t), &release.Pilot{FlagID: newUUID(), EnvironmentID: sh.ExampleDeploymentEnvironment(t).ID, PublicID: extID, IsParticipating: false})
				})

				s.Then(`it will return an empty result set`, func(t *testcase.T) {
					count, err := iterators.Count(subject(t))
					require.Nil(t, err)
					require.Equal(t, 0, count)
				})
			})

			s.When(`pilot ext id has multiple records`, func(s *testcase.Spec) {
				expectedPilots := s.Let(`expected pilots`, func(t *testcase.T) interface{} {
					var pilots []release.Pilot

					for i := 0; i < rand.Intn(5)+5; i++ {
						uuidV4, err := uuid.NewRandom()
						require.Nil(t, err)

						pilot := release.Pilot{
							FlagID:          uuidV4.String(),
							EnvironmentID:   sh.ExampleDeploymentEnvironment(t).ID,
							PublicID:        sh.ExampleExternalPilotID(t),
							IsParticipating: rand.Intn(1) == 0,
						}

						contracts.CreateEntity(t, c.storageGet(t), c.context(t), &pilot)
						pilots = append(pilots, pilot)
					}

					return pilots
				}).EagerLoading(s)

				s.Then(`it will return all of them`, func(t *testcase.T) {
					var pilots []release.Pilot
					require.Nil(t, iterators.Collect(subject(t), &pilots))
					require.ElementsMatch(t, expectedPilots.Get(t).([]release.Pilot), pilots)
				})
			})
		})
	})
}
