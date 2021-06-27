package spechelper

import (
	"fmt"
	"net/url"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/testcase"
	"github.com/google/uuid"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
)

type FixtureFactory struct {
	fixtures.FixtureFactory
}

func (ff FixtureFactory) Create(EntityType interface{}) interface{} {
	switch reflects.BaseValueOf(EntityType).Interface().(type) {
	case release.Flag:
		flag := ff.FixtureFactory.Create(EntityType).(*release.Flag)
		flag.Name = fmt.Sprintf(`%s - %s`, flag.Name, uuid.New().String())
		return flag

	case release.RolloutDecisionByAPI:
		byAPI := release.NewRolloutDecisionByAPI()
		u, err := url.ParseRequestURI(fmt.Sprintf(`https://example.com//%s`, url.PathEscape(fixtures.Random.String())))
		if err != nil {
			panic(err.Error())
		}
		byAPI.URL = u
		return &byAPI

	case release.RolloutDecisionByPercentage:
		r := release.NewRolloutDecisionByPercentage()
		r.Percentage = fixtures.Random.IntBetween(0, 100)
		r.Seed = int64(fixtures.Random.IntBetween(0, 1024))
		return &r

	case security.Token:
		t := ff.FixtureFactory.Create(EntityType).(*security.Token)
		t.SHA512 = uuid.New().String()
		return t

	default:
		return ff.FixtureFactory.Create(EntityType)
	}
}

func (ff FixtureFactory) Dynamic(t *testcase.T) DynamicFixtureFactory {
	return DynamicFixtureFactory{
		T:              t,
		FixtureFactory: ff,
	}
}

type DynamicFixtureFactory struct {
	T *testcase.T
	FixtureFactory
}

func (ff DynamicFixtureFactory) Create(T interface{}) interface{} {
	switch reflects.BaseValueOf(T).Interface().(type) {
	case release.Pilot:
		pilot := ff.FixtureFactory.Create(T).(*release.Pilot)
		pilot.PublicID = uuid.New().String()
		pilot.FlagID = ExampleReleaseFlag(ff.T).ID
		pilot.EnvironmentID = ExampleDeploymentEnvironment(ff.T).ID
		return pilot

	case release.Rollout:
		r := ff.FixtureFactory.Create(T).(*release.Rollout)
		r.DeploymentEnvironmentID = ExampleDeploymentEnvironment(ff.T).ID
		r.FlagID = ExampleReleaseFlag(ff.T).ID
		r.Plan = func() release.RolloutDefinition {
			switch fixtures.Random.IntN(3) {
			case 0:
				byPercentage := release.NewRolloutDecisionByPercentage()
				byPercentage.Percentage = fixtures.Random.IntBetween(0, 100)
				return byPercentage

			case 1:
				byAPI := release.NewRolloutDecisionByAPI()
				u, err := url.ParseRequestURI(fmt.Sprintf(`https://example.com/%s`, url.PathEscape(fixtures.Random.String())))
				if err != nil {
					panic(err.Error())
				}
				byAPI.URL = u
				return byAPI

			case 2:
				byPercentage := release.NewRolloutDecisionByPercentage()
				byPercentage.Percentage = fixtures.Random.IntBetween(0, 100)

				byAPI := release.NewRolloutDecisionByAPI()
				u, err := url.ParseRequestURI(fmt.Sprintf(`https://example.com/%s`, url.PathEscape(fixtures.Random.String())))
				if err != nil {
					panic(err.Error())
				}
				byAPI.URL = u

				return release.RolloutDecisionAND{
					Left:  byPercentage,
					Right: byAPI,
				}

			default:
				panic(`shouldn't be the case`)
			}
		}()
		return r

	default:
		return ff.FixtureFactory.Create(T)
	}
}

var DefaultFixtureFactory = FixtureFactory{}

func Create(T interface{}) (ptr interface{}) {
	return DefaultFixtureFactory.Create(T)
}
