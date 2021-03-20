package contracts

import (
	"context"
	"math/rand"
	"strconv"
	"testing"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/frameless/resources/contracts"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	sh "github.com/toggler-io/toggler/spechelper"
)

type ManualPilotStorage struct {
	Subject        func(testing.TB) ManualPilotStorageSubject
	FixtureFactory sh.FixtureFactory
}

type ManualPilotStorageSubject interface {
	resources.Creator
	resources.Finder
	resources.Updater
	resources.Deleter
	resources.OnePhaseCommitProtocol
	resources.CreatorPublisher
	resources.UpdaterPublisher
	resources.DeleterPublisher
	release.PilotFinder
}

func (spec ManualPilotStorage) storage() testcase.Var {
	return testcase.Var{
		Name: "release pilot storage",
		Init: func(t *testcase.T) interface{} {
			return spec.Subject(t)
		},
	}
}

func (spec ManualPilotStorage) storageGet(t *testcase.T) ManualPilotStorageSubject {
	return spec.storage().Get(t).(ManualPilotStorageSubject)
}

func (spec ManualPilotStorage) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec ManualPilotStorage) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec ManualPilotStorage) setUp(s *testcase.Spec) {
	sh.SetUp(s)

	sh.Storage.Let(s, func(t *testcase.T) interface{} {
		return spec.storageGet(t)
	})

	s.After(func(t *testcase.T) {
		// TODO: delete this?
		require.Nil(t, spec.storageGet(t).DeleteAll(spec.FixtureFactory.Context(), release.Flag{}))
		require.Nil(t, spec.storageGet(t).DeleteAll(spec.FixtureFactory.Context(), release.ManualPilot{}))
	})
}

func (spec ManualPilotStorage) Spec(tb testing.TB) {
	testcase.NewSpec(tb).Describe(`ManualPilotStorage`, func(s *testcase.Spec) {
		spec.setUp(s)

		s.Test(`contracts`, func(t *testcase.T) {
			T := release.ManualPilot{}
			testcase.RunContract(t.TB,
				contracts.OnePhaseCommitProtocol{T: T,
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
					Subject: func(tb testing.TB) contracts.OnePhaseCommitProtocolSubject {
						return spec.Subject(tb)
					},
				},
				contracts.Creator{T: T,
					Subject: func(tb testing.TB) contracts.CRD {
						return spec.Subject(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
				contracts.Finder{T: T,
					Subject: func(tb testing.TB) contracts.CRD {
						return spec.Subject(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
				contracts.Updater{T: T,
					Subject: func(tb testing.TB) contracts.UpdaterSubject {
						return spec.Subject(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
				contracts.Deleter{T: T,
					Subject: func(tb testing.TB) contracts.CRD {
						return spec.Subject(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
				ManualPilotFinder{
					FixtureFactory: spec.FixtureFactory,
					Subject: func(tb testing.TB) ManualPilotFinderSubject {
						return spec.Subject(tb)
					},
				},
				contracts.CreatorPublisher{T: T,
					Subject: func(tb testing.TB) contracts.CreatorPublisherSubject {
						return spec.Subject(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
				contracts.UpdaterPublisher{T: T,
					Subject: func(tb testing.TB) contracts.UpdaterPublisherSubject {
						return spec.Subject(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
				contracts.DeleterPublisher{T: T,
					Subject: func(tb testing.TB) contracts.DeleterPublisherSubject {
						return spec.Subject(tb)
					},
					FixtureFactory: spec.FixtureFactory.Dynamic(t),
				},
			)
		})
	})
}

///////////////////////////////////////////////////////- query -////////////////////////////////////////////////////////

type ManualPilotFinder struct {
	Subject        func(testing.TB) ManualPilotFinderSubject
	FixtureFactory sh.FixtureFactory
}

type ManualPilotFinderSubject interface {
	release.PilotFinder
	resources.Creator
	resources.Updater
	resources.Finder
	resources.Deleter
	resources.OnePhaseCommitProtocol
}

func (spec ManualPilotFinder) storage() testcase.Var {
	return testcase.Var{
		Name: "release pilot storage",
		Init: func(t *testcase.T) interface{} {
			return spec.Subject(t)
		},
	}
}

func (spec ManualPilotFinder) storageGet(t *testcase.T) ManualPilotFinderSubject {
	return spec.storage().Get(t).(ManualPilotFinderSubject)
}

func (spec ManualPilotFinder) Test(t *testing.T) {
	s := testcase.NewSpec(t)

	sh.SetUp(s)

	sh.Storage.Let(s, func(t *testcase.T) interface{} {
		return spec.storageGet(t)
	})

	s.Describe(`ManualPilotFinder`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			contracts.DeleteAllEntity(t, spec.storageGet(t), spec.context(), release.ManualPilot{})
		})
		s.After(func(t *testcase.T) {
			contracts.DeleteAllEntity(t, spec.storageGet(t), spec.context(), release.ManualPilot{})
		})

		s.Describe(`FindReleasePilotsByReleaseFlag`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) iterators.Interface {
				pilotEntriesIter := spec.storageGet(t).FindReleasePilotsByReleaseFlag(spec.context(), *sh.ExampleReleaseFlag(t))
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
					return spec.FixtureFactory.Create(release.Flag{})
				})

				thenNoPilotsFound(s)
			})

			s.When(`flag is persisted`, func(s *testcase.Spec) {
				thenNoPilotsFound(s)

				s.And(`there are manual pilot configs for the release flag`, func(s *testcase.Spec) {
					expectedPilots := s.Let(`expectedPilots`, func(t *testcase.T) interface{} {
						var expectedPilots []*release.ManualPilot
						for i := 0; i < 5; i++ {
							pilot := &release.ManualPilot{
								FlagID:                  sh.ExampleReleaseFlag(t).ID,
								DeploymentEnvironmentID: sh.ExampleDeploymentEnvironment(t).ID,
								ExternalID:              strconv.Itoa(i),
							}

							contracts.CreateEntity(t, spec.storageGet(t), spec.context(), pilot)
							expectedPilots = append(expectedPilots, pilot)
						}
						return expectedPilots
					}).EagerLoading(s)

					expectedPilotsGet := func(t *testcase.T) []*release.ManualPilot {
						return expectedPilots.Get(t).([]*release.ManualPilot)
					}

					s.Then(`it will return all of them`, func(t *testcase.T) {
						iter := subject(t)
						defer iter.Close()
						require.NotNil(t, iter)
						var actualPilots []*release.ManualPilot
						for iter.Next() {
							var actually release.ManualPilot
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
			var subject = func(t *testcase.T) (*release.ManualPilot, error) {
				return spec.storageGet(t).FindReleaseManualPilotByExternalID(
					spec.context(),
					sh.ExampleReleaseFlag(t).ID,
					sh.ExampleDeploymentEnvironment(t).ID,
					sh.ExampleID(t),
				)
			}

			s.Before(func(t *testcase.T) {
				contracts.DeleteAllEntity(t, spec.storageGet(t), spec.context(), release.ManualPilot{})
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
					return spec.FixtureFactory.Create(release.Flag{})
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
						pilot := &release.ManualPilot{
							FlagID:                  sh.ExampleReleaseFlag(t).ID,
							DeploymentEnvironmentID: sh.ExampleDeploymentEnvironment(t).ID,
							ExternalID:              sh.ExampleID(t),
						}
						contracts.CreateEntity(t, spec.storageGet(t), spec.context(), pilot)
					})

					s.Then(`then pilots will be retrieved`, func(t *testcase.T) {
						pilot, err := subject(t)
						require.Nil(t, err)
						require.NotNil(t, pilot)

						require.Equal(t, sh.ExampleID(t), pilot.ExternalID)
						require.Equal(t, sh.ExampleReleaseFlag(t).ID, pilot.FlagID)
						require.Equal(t, sh.ExampleDeploymentEnvironment(t).ID, pilot.DeploymentEnvironmentID)
					})
				})
			})
		})

		s.Describe(`FindReleasePilotsByExternalID`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) iterators.Interface {
				pilotEntriesIter := spec.storageGet(t).FindReleasePilotsByExternalID(spec.context(), sh.ExampleExternalPilotID(t))
				t.Defer(pilotEntriesIter.Close)
				return pilotEntriesIter
			}

			s.Let(`PilotExternalID`, func(t *testcase.T) interface{} {
				return fixtures.Random.String()
			})

			s.When(`there is no pilot records`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					contracts.DeleteAllEntity(t, spec.storageGet(t), spec.context(), release.ManualPilot{})
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

					contracts.CreateEntity(t, spec.storageGet(t), spec.context(), &release.ManualPilot{FlagID: newUUID(), DeploymentEnvironmentID: sh.ExampleDeploymentEnvironment(t).ID, ExternalID: extID, IsParticipating: true})
					contracts.CreateEntity(t, spec.storageGet(t), spec.context(), &release.ManualPilot{FlagID: newUUID(), DeploymentEnvironmentID: sh.ExampleDeploymentEnvironment(t).ID, ExternalID: extID, IsParticipating: true})
					contracts.CreateEntity(t, spec.storageGet(t), spec.context(), &release.ManualPilot{FlagID: newUUID(), DeploymentEnvironmentID: sh.ExampleDeploymentEnvironment(t).ID, ExternalID: extID, IsParticipating: false})
				})

				s.Then(`it will return an empty result set`, func(t *testcase.T) {
					count, err := iterators.Count(subject(t))
					require.Nil(t, err)
					require.Equal(t, 0, count)
				})
			})

			s.When(`pilot ext id has multiple records`, func(s *testcase.Spec) {
				expectedPilots := s.Let(`expected pilots`, func(t *testcase.T) interface{} {
					var pilots []release.ManualPilot

					for i := 0; i < rand.Intn(5)+5; i++ {
						uuidV4, err := uuid.NewRandom()
						require.Nil(t, err)

						pilot := release.ManualPilot{
							FlagID:                  uuidV4.String(),
							DeploymentEnvironmentID: sh.ExampleDeploymentEnvironment(t).ID,
							ExternalID:              sh.ExampleExternalPilotID(t),
							IsParticipating:         rand.Intn(1) == 0,
						}

						contracts.CreateEntity(t, spec.storageGet(t), spec.context(), &pilot)
						pilots = append(pilots, pilot)
					}

					return pilots
				}).EagerLoading(s)

				s.Then(`it will return all of them`, func(t *testcase.T) {
					var pilots []release.ManualPilot
					require.Nil(t, iterators.Collect(subject(t), &pilots))
					require.ElementsMatch(t, expectedPilots.Get(t).([]release.ManualPilot), pilots)
				})
			})
		})
	})
}

func (spec ManualPilotFinder) Benchmark(b *testing.B) {
	b.Run(`ManualPilotFinder`, func(b *testing.B) {
		b.Skip(`TODO`)
	})
}

func (spec ManualPilotFinder) context() context.Context {
	return spec.FixtureFactory.Context()
}
