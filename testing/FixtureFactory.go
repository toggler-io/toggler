package testing

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/google/uuid"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
)

func NewFixtureFactory() *FixtureFactory {
	return &FixtureFactory{}
}

type FixtureFactory struct {
	fixtures.FixtureFactory
}

// this ensures that the randoms have better variety between test runs with -count n
var rnd = rand.New(rand.NewSource(time.Now().Unix()))

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

	case release.ManualPilot:
		pilot := ff.FixtureFactory.Create(EntityType).(*release.ManualPilot)
		pilot.ExternalID = uuid.New().String()
		return pilot

	case security.Token:
		t := ff.FixtureFactory.Create(EntityType).(*security.Token)
		t.SHA512 = uuid.New().String()
		return t

	default:
		return ff.FixtureFactory.Create(EntityType)
	}
}

var DefaultFixtureFactory = FixtureFactory{}

func Create(T interface{}) (ptr interface{}) {
	return DefaultFixtureFactory.Create(T)
}
