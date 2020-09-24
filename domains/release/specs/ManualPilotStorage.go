package specs

import (
	"context"
	"math/rand"
	"strconv"
	"testing"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	. "github.com/toggler-io/toggler/testing"
)

type ManualPilotStorage struct {
	Subject interface {
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

	FixtureFactory FixtureFactory
}

func (spec ManualPilotStorage) Test(t *testing.T) {
	t.Run(`ManualPilotStorage`, func(t *testing.T) {
		spec.Spec(t)
	})
}

func (spec ManualPilotStorage) Benchmark(b *testing.B) {
	b.Run(`ManualPilotStorage`, func(b *testing.B) {
		spec.Spec(b)
	})
}

func (spec ManualPilotStorage) setup(s *testcase.Spec) {
	SetUp(s)

	s.Let(LetVarExampleStorage, func(t *testcase.T) interface{} {
		return spec.Subject
	})

	s.After(func(t *testcase.T) {
		require.Nil(t, spec.Subject.DeleteAll(spec.FixtureFactory.Context(), release.Flag{}))
		require.Nil(t, spec.Subject.DeleteAll(spec.FixtureFactory.Context(), release.ManualPilot{}))
	})
}

func (spec ManualPilotStorage) Spec(tb testing.TB) {
	s := testcase.NewSpec(tb)
	spec.setup(s)

	s.Test(``, func(t *testcase.T) {
		specs.Run(t.TB,
			specs.OnePhaseCommitProtocol{
				EntityType:     release.ManualPilot{},
				FixtureFactory: spec.FixtureFactory.Dynamic(t),
				Subject:        spec.Subject,
			},
			specs.CRUD{
				EntityType:     release.ManualPilot{},
				FixtureFactory: spec.FixtureFactory.Dynamic(t),
				Subject:        spec.Subject,
			},
			ManualPilotFinder{
				FixtureFactory: spec.FixtureFactory,
				Subject:        spec.Subject,
			},
			specs.CreatorPublisher{
				Subject:        spec.Subject,
				EntityType:     release.ManualPilot{},
				FixtureFactory: spec.FixtureFactory.Dynamic(t),
			},
			specs.UpdaterPublisher{
				Subject:        spec.Subject,
				EntityType:     release.ManualPilot{},
				FixtureFactory: spec.FixtureFactory.Dynamic(t),
			},
			specs.DeleterPublisher{
				Subject:        spec.Subject,
				EntityType:     release.ManualPilot{},
				FixtureFactory: spec.FixtureFactory.Dynamic(t),
			},
		)
	})
}

///////////////////////////////////////////////////////- query -////////////////////////////////////////////////////////

type ManualPilotFinder struct {
	Subject interface {
		release.PilotFinder
		resources.Creator
		resources.Updater
		resources.Finder
		resources.Deleter
		resources.OnePhaseCommitProtocol
	}

	FixtureFactory FixtureFactory
}

func (spec ManualPilotFinder) Test(t *testing.T) {
	s := testcase.NewSpec(t)
	SetUp(s)

	s.Let(LetVarExampleStorage, func(t *testcase.T) interface{} {
		return spec.Subject
	})

	s.Describe(`ManualPilotFinder`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			require.Nil(t, spec.Subject.DeleteAll(spec.context(), release.ManualPilot{}))
		})

		s.After(func(t *testcase.T) {
			require.Nil(t, spec.Subject.DeleteAll(spec.context(), release.ManualPilot{}))
		})

		s.Describe(`FindReleasePilotsByReleaseFlag`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) iterators.Interface {
				pilotEntriesIter := spec.Subject.FindReleasePilotsByReleaseFlag(spec.context(), *ExampleReleaseFlag(t))
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
				s.Let(LetVarExampleReleaseFlag, func(t *testcase.T) interface{} {
					return spec.FixtureFactory.Create(release.Flag{})
				})

				thenNoPilotsFound(s)
			})

			s.When(`flag is persisted`, func(s *testcase.Spec) {
				thenNoPilotsFound(s)

				s.And(`there are manual pilot configs for the release flag`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						expectedPilots := t.I(`expectedPilots`).([]*release.ManualPilot)

						for _, pilot := range expectedPilots {
							require.Nil(t, spec.Subject.Create(spec.context(), pilot))
						}
					})

					s.Let(`expectedPilots`, func(t *testcase.T) interface{} {
						var expectedPilots []*release.ManualPilot
						for i := 0; i < 5; i++ {
							pilot := &release.ManualPilot{
								FlagID:                  ExampleReleaseFlag(t).ID,
								DeploymentEnvironmentID: ExampleDeploymentEnvironment(t).ID,
								ExternalID:              strconv.Itoa(i),
							}

							expectedPilots = append(expectedPilots, pilot)
						}
						return expectedPilots
					})

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

						expectedPilots := t.I(`expectedPilots`).([]*release.ManualPilot)
						require.True(t, len(expectedPilots) == len(actualPilots))
						require.ElementsMatch(t, expectedPilots, actualPilots)
					})
				})
			})
		})

		s.Describe(`FindReleaseManualPilotByExternalID`, func(s *testcase.Spec) {
			var subject = func(t *testcase.T) (*release.ManualPilot, error) {
				return spec.Subject.FindReleaseManualPilotByExternalID(
					spec.context(),
					ExampleReleaseFlag(t).ID,
					ExampleDeploymentEnvironment(t).ID,
					ExampleID(t),
				)
			}

			s.Before(func(t *testcase.T) {
				require.Nil(t, spec.Subject.DeleteAll(spec.context(), release.ManualPilot{}))
			})

			ThenNoPilotsFound := func(s *testcase.Spec) {
				s.Then(`no pilots found`, func(t *testcase.T) {
					pilot, err := subject(t)
					require.Nil(t, err)
					require.Nil(t, pilot)
				})
			}

			s.When(`flag is not persisted`, func(s *testcase.Spec) {
				s.Let(LetVarExampleReleaseFlag, func(t *testcase.T) interface{} {
					return spec.FixtureFactory.Create(release.Flag{})
				})

				ThenNoPilotsFound(s)
			})

			s.When(`flag persisted already exists`, func(s *testcase.Spec) {
				s.Let(`featureFlagID`, func(t *testcase.T) interface{} {
					return ExampleReleaseFlag(t).ID
				})

				ThenNoPilotsFound(s)

				s.And(`the given there is a registered pilot for the feature`, func(s *testcase.Spec) {
					s.Around(func(t *testcase.T) func() {
						pilot := &release.ManualPilot{
							FlagID:                  ExampleReleaseFlag(t).ID,
							DeploymentEnvironmentID: ExampleDeploymentEnvironment(t).ID,
							ExternalID:              ExampleID(t),
						}
						require.Nil(t, spec.Subject.Create(spec.context(), pilot))
						return func() { require.Nil(t, spec.Subject.DeleteByID(spec.context(), *pilot, pilot.ID)) }
					})

					s.Then(`then pilots will be retrieved`, func(t *testcase.T) {
						pilot, err := subject(t)
						require.Nil(t, err)
						require.NotNil(t, pilot)

						require.Equal(t, ExampleID(t), pilot.ExternalID)
						require.Equal(t, ExampleReleaseFlag(t).ID, pilot.FlagID)
						require.Equal(t, ExampleDeploymentEnvironment(t).ID, pilot.DeploymentEnvironmentID)
					})
				})
			})
		})

		s.Describe(`FindReleasePilotsByExternalID`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) iterators.Interface {
				pilotEntriesIter := spec.Subject.FindReleasePilotsByExternalID(spec.context(), ExampleExternalPilotID(t))
				t.Defer(pilotEntriesIter.Close)
				return pilotEntriesIter
			}

			s.Let(`PilotExternalID`, func(t *testcase.T) interface{} {
				return fixtures.Random.String()
			})

			s.When(`there is no pilot records`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) { require.Nil(t, spec.Subject.DeleteAll(spec.context(), release.ManualPilot{})) })

				s.Then(`it will return an empty result set`, func(t *testcase.T) {
					count, err := iterators.Count(subject(t))
					require.Nil(t, err)
					require.Equal(t, 0, count)
				})
			})

			s.When(`the given pilot id has no records`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					ctx := spec.context()
					extID := fixtures.Random.String()

					var newUUID = func() string {
						uuidV4, err := uuid.NewRandom()
						require.Nil(t, err)
						return uuidV4.String()
					}

					require.Nil(t, spec.Subject.Create(ctx, &release.ManualPilot{FlagID: newUUID(), DeploymentEnvironmentID: ExampleDeploymentEnvironment(t).ID, ExternalID: extID, IsParticipating: true}))
					require.Nil(t, spec.Subject.Create(ctx, &release.ManualPilot{FlagID: newUUID(), DeploymentEnvironmentID: ExampleDeploymentEnvironment(t).ID, ExternalID: extID, IsParticipating: true}))
					require.Nil(t, spec.Subject.Create(ctx, &release.ManualPilot{FlagID: newUUID(), DeploymentEnvironmentID: ExampleDeploymentEnvironment(t).ID, ExternalID: extID, IsParticipating: false}))
				})

				s.Then(`it will return an empty result set`, func(t *testcase.T) {
					count, err := iterators.Count(subject(t))
					require.Nil(t, err)
					require.Equal(t, 0, count)
				})
			})

			s.When(`pilot ext id has multiple records`, func(s *testcase.Spec) {
				s.Let(`expected pilots`, func(t *testcase.T) interface{} {
					var pilots []release.ManualPilot

					for i := 0; i < rand.Intn(5)+5; i++ {
						uuidV4, err := uuid.NewRandom()
						require.Nil(t, err)

						pilot := release.ManualPilot{
							FlagID:                  uuidV4.String(),
							DeploymentEnvironmentID: ExampleDeploymentEnvironment(t).ID,
							ExternalID:              ExampleExternalPilotID(t),
							IsParticipating:         rand.Intn(1) == 0,
						}

						require.Nil(t, spec.Subject.Create(spec.context(), &pilot))
						pilots = append(pilots, pilot)
					}

					return pilots
				})

				s.Before(func(t *testcase.T) { t.I(`expected pilots`) }) // eager load let value

				s.Then(`it will return all of them`, func(t *testcase.T) {
					var pilots []release.ManualPilot
					require.Nil(t, iterators.Collect(subject(t), &pilots))
					require.ElementsMatch(t, t.I(`expected pilots`).([]release.ManualPilot), pilots)
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
