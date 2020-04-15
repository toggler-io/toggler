package specs

import (
	"context"
	"math/rand"
	"strconv"
	"testing"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/testcase"
	"github.com/google/uuid"

	. "github.com/toggler-io/toggler/testing"

	"github.com/toggler-io/toggler/domains/release"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/stretchr/testify/require"
)

type pilotFinderSpec struct {
	Subject interface {
		release.PilotFinder

		specs.MinimumRequirements
	}

	specs.FixtureFactory
}

func (spec pilotFinderSpec) Benchmark(b *testing.B) {
	b.Run(`pilotFinderSpec`, func(b *testing.B) {
		b.Skip(`TODO`)

		b.Run(`FindPilotsByFeatureFlag`, func(b *testing.B) {
			flag := spec.Create(release.Flag{}).(*release.Flag)
			require.Nil(b, spec.Subject.Create(spec.Context(), flag))
			pilots := CreateEntities(specs.BenchmarkEntityVolumeCount(), spec.FixtureFactory, release.Pilot{})
			SaveEntities(b, spec.Subject, spec.FixtureFactory, pilots...)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := iterators.Count(spec.Subject.FindPilotsByFeatureFlag(spec.Context(), flag))
				require.Nil(b, err)
			}
		})
	})
}

func (spec pilotFinderSpec) Test(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Describe(`pilotFinderSpec`, func(s *testcase.Spec) {

		s.Let(`flagName`, func(t *testcase.T) interface{} {
			return RandomName()
		})

		s.Before(func(t *testcase.T) {
			require.Nil(t, spec.Subject.DeleteAll(spec.ctx(), release.Flag{}))
			require.Nil(t, spec.Subject.DeleteAll(spec.ctx(), release.Pilot{}))
		})

		s.After(func(t *testcase.T) {
			require.Nil(t, spec.Subject.DeleteAll(spec.ctx(), release.Flag{}))
			require.Nil(t, spec.Subject.DeleteAll(spec.ctx(), release.Pilot{}))
		})

		s.Describe(`FindPilotsByFeatureFlag`, func(s *testcase.Spec) {
			getFF := func(t *testcase.T) *release.Flag {
				var ff *release.Flag
				f := t.I(`ff`)
				if f != nil {
					ff = f.(*release.Flag)
				}
				return ff
			}

			subject := func(t *testcase.T) frameless.Iterator {
				pilotEntriesIter := spec.Subject.FindPilotsByFeatureFlag(spec.ctx(), getFF(t))
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

			s.When(`feature object is nil`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.DeleteAll(spec.ctx(), release.Flag{}))
				})
				s.Let(`ff`, func(t *testcase.T) interface{} { return nil })
				thenNoPilotsFound(s)
			})

			s.When(`feature object has no reference`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.DeleteAll(spec.ctx(), release.Flag{}))
				})
				s.Let(`ff`, func(t *testcase.T) interface{} { return &release.Flag{} })
				thenNoPilotsFound(s)
			})

			s.When(`feature flag exists`, func(s *testcase.Spec) {
				s.Let(`ff`, func(t *testcase.T) interface{} {
					ff := &release.Flag{Name: t.I(`flagName`).(string)}
					require.Nil(t, spec.Subject.Create(spec.ctx(), ff))
					return ff
				})

				thenNoPilotsFound(s)

				s.And(`there are registered pilots for the feature`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						expectedPilots := t.I(`expectedPilots`).([]*release.Pilot)

						for _, pilot := range expectedPilots {
							require.Nil(t, spec.Subject.Create(spec.ctx(), pilot))
						}
					})

					s.Let(`expectedPilots`, func(t *testcase.T) interface{} {
						var expectedPilots []*release.Pilot
						ff := t.I(`ff`).(*release.Flag)

						for i := 0; i < 5; i++ {
							pilot := &release.Pilot{FlagID: ff.ID, ExternalID: strconv.Itoa(i)}
							expectedPilots = append(expectedPilots, pilot)
						}

						return expectedPilots
					})

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

						expectedPilots := t.I(`expectedPilots`).([]*release.Pilot)

						require.True(t, len(expectedPilots) == len(actualPilots))

						for _, expected := range expectedPilots {
							require.Contains(t, actualPilots, expected)
						}
					})
				})
			})
		})

		s.Describe(`FindReleaseFlagPilotByPilotExternalID`, func(s *testcase.Spec) {
			const ExternalPublicPilotID = `42`

			subject := func(t *testcase.T) (*release.Pilot, error) {
				return spec.Subject.FindReleaseFlagPilotByPilotExternalID(spec.ctx(), t.I(`featureFlagID`).(string), ExternalPublicPilotID)
			}

			ThenNoPilotsFound := func(s *testcase.Spec) {
				s.Then(`no pilots found`, func(t *testcase.T) {
					pilot, err := subject(t)
					require.Nil(t, err)
					require.Nil(t, pilot)
				})
			}

			s.When(`feature was never enabled before`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					require.Nil(t, spec.Subject.DeleteAll(spec.ctx(), release.Flag{}))
				})
				s.Let(`featureFlagID`, func(t *testcase.T) interface{} { return "not exinsting ID" })
				ThenNoPilotsFound(s)
			})

			s.When(`feature flag exists`, func(s *testcase.Spec) {
				s.Let(`featureFlagID`, func(t *testcase.T) interface{} {
					ff := &release.Flag{Name: t.I(`flagName`).(string)}
					ff.Rollout.Strategy.Percentage = 100
					require.Nil(t, spec.Subject.Create(spec.ctx(), ff))
					return ff.ID
				})

				ThenNoPilotsFound(s)

				s.And(`the given there is a registered pilot for the feature`, func(s *testcase.Spec) {
					s.Before(func(t *testcase.T) {
						require.Nil(t, spec.Subject.DeleteAll(spec.ctx(), release.Pilot{}))
						featureFlagID := t.I(`featureFlagID`).(string)
						pilot := &release.Pilot{FlagID: featureFlagID, ExternalID: ExternalPublicPilotID}
						require.Nil(t, spec.Subject.Create(spec.ctx(), pilot))
					})

					s.Then(`asd`, func(t *testcase.T) {
						pilot, err := subject(t)
						require.Nil(t, err)
						require.NotNil(t, pilot)

						featureFlagID := t.I(`featureFlagID`).(string)
						require.Equal(t, ExternalPublicPilotID, pilot.ExternalID)
						require.Equal(t, featureFlagID, pilot.FlagID)
					})
				})
			})
		})

		s.Describe(`FindPilotEntriesByExtID`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) frameless.Iterator {
				pilotEntriesIter := spec.Subject.FindPilotEntriesByExtID(spec.ctx(), GetExternalPilotID(t))
				t.Defer(pilotEntriesIter.Close)
				return pilotEntriesIter
			}

			s.Let(`PilotExternalID`, func(t *testcase.T) interface{} {
				return RandomExternalPilotID()
			})

			s.When(`there is no pilot records`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) { require.Nil(t, spec.Subject.DeleteAll(spec.ctx(), release.Pilot{})) })

				s.Then(`it will return an empty result set`, func(t *testcase.T) {
					count, err := iterators.Count(subject(t))
					require.Nil(t, err)
					require.Equal(t, 0, count)
				})
			})

			s.When(`the given pilot id has no records`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					ctx := spec.ctx()
					extID := RandomExternalPilotID()

					var newUUID = func() string {
						uuidV4, err := uuid.NewRandom()
						require.Nil(t, err)
						return uuidV4.String()
					}

					require.Nil(t, spec.Subject.Create(ctx, &release.Pilot{FlagID: newUUID(), ExternalID: extID, Enrolled: true}))
					require.Nil(t, spec.Subject.Create(ctx, &release.Pilot{FlagID: newUUID(), ExternalID: extID, Enrolled: true}))
					require.Nil(t, spec.Subject.Create(ctx, &release.Pilot{FlagID: newUUID(), ExternalID: extID, Enrolled: false}))
				})

				s.Then(`it will return an empty result set`, func(t *testcase.T) {
					count, err := iterators.Count(subject(t))
					require.Nil(t, err)
					require.Equal(t, 0, count)
				})
			})

			s.When(`pilot ext id has multiple records`, func(s *testcase.Spec) {
				s.Let(`expected pilots`, func(t *testcase.T) interface{} {
					var pilots []release.Pilot

					for i := 0; i < rand.Intn(5)+5; i++ {
						uuidV4, err := uuid.NewRandom()
						require.Nil(t, err)

						pilot := release.Pilot{
							FlagID:     uuidV4.String(),
							ExternalID: GetExternalPilotID(t),
							Enrolled:   rand.Intn(1) == 0,
						}

						require.Nil(t, spec.Subject.Create(spec.ctx(), &pilot))
						pilots = append(pilots, pilot)
					}

					return pilots
				})

				s.Before(func(t *testcase.T) { t.I(`expected pilots`) }) // eager load let value

				s.Then(`it will return all of them`, func(t *testcase.T) {
					var pilots []release.Pilot
					require.Nil(t, iterators.Collect(subject(t), &pilots))
					require.ElementsMatch(t, t.I(`expected pilots`).([]release.Pilot), pilots)
				})
			})

		})
	})
}

func (spec pilotFinderSpec) ctx() context.Context {
	return spec.FixtureFactory.Context()
}
