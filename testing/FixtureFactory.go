package testing

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/frameless/resources/specs"
	"github.com/adamluzsi/toggler/services/rollouts"
)

func NewFixtureFactory() *FixtureFactory {
	return &FixtureFactory{}
}

type FixtureFactory struct {
	specs.GenericFixtureFactory
	PilotFeatureFlagID string // this will allow to create pilot fixtures
}

// this ensures that the randoms have better variety between test runs with -count n
var rnd = rand.New(rand.NewSource(time.Now().Unix()))

func (ff *FixtureFactory) SetPilotFeatureFlagID(ffID string) func() {
	original := ff.PilotFeatureFlagID
	ff.PilotFeatureFlagID = ffID
	return func() { ff.PilotFeatureFlagID = original }
}

func (ff *FixtureFactory) Create(EntityType interface{}) interface{} {
	switch EntityType.(type) {
	case rollouts.FeatureFlag, *rollouts.FeatureFlag:
		flag := ff.GenericFixtureFactory.Create(EntityType).(*rollouts.FeatureFlag)

		flag.Rollout.Strategy.DecisionLogicAPI = nil

		if rnd.Intn(2) == 0 {
			u, err := url.ParseRequestURI(fmt.Sprintf(`http://google.com/%s`, url.PathEscape(fixtures.RandomString(13))))

			if err != nil {
				panic(err)
			}

			flag.Rollout.Strategy.DecisionLogicAPI = u
		}

		return flag

	case rollouts.Pilot, *rollouts.Pilot:
		pilot := ff.GenericFixtureFactory.Create(EntityType).(*rollouts.Pilot)
		pilot.FeatureFlagID = ff.PilotFeatureFlagID
		return pilot

	default:
		return ff.GenericFixtureFactory.Create(EntityType)
	}
}
