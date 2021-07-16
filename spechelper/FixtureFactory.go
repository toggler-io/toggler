package spechelper

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/adamluzsi/frameless/contracts"
	"net/url"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/google/uuid"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
)

var FixtureFactory = testcase.Var{Name: "contracts.FixtureFactory"}

func FixtureFactoryGet(t *testcase.T) contracts.FixtureFactory {
	return FixtureFactory.Get(t).(contracts.FixtureFactory)
}

func FixtureFactoryLet(s *testcase.Spec, blk func(testing.TB) contracts.FixtureFactory) {
	FixtureFactory.Let(s, func(t *testcase.T) interface{} { return blk(t) })
}

func NewFixtureFactory(tb testing.TB) *fixtures.Factory {
	t, ok := tb.(*testcase.T)
	if !ok {
		t = testcase.NewT(tb, NewSpec(tb))
	}
	factory := &fixtures.Factory{
		Random:  t.Random,
		Options: []fixtures.Option{fixtures.SkipByTag(`ext`, "id", "ID")},
	}
	factory.RegisterType(release.Flag{}, func() interface{} {
		return release.Flag{
			Name: fmt.Sprintf(`%s - %s`, t.Random.StringN(4), uuid.New().String()),
		}
	})
	factory.RegisterType(release.RolloutDecisionByAPI{}, func() interface{} {
		var byAPI release.RolloutDecisionByAPI
		byAPI = release.NewRolloutDecisionByAPI()
		u, err := url.ParseRequestURI(fmt.Sprintf(`https://example.com/%s`, url.PathEscape(t.Random.String())))
		if err != nil {
			panic(err.Error())
		}
		byAPI.URL = u
		return byAPI
	})
	factory.RegisterType(release.RolloutDecisionByPercentage{}, func() interface{} {
		var r release.RolloutDecisionByPercentage
		r = release.NewRolloutDecisionByPercentage()
		r.Percentage = t.Random.IntBetween(0, 100)
		r.Seed = int64(t.Random.IntBetween(0, 1024))
		return r
	})
	factory.RegisterType(security.Token{}, func() interface{} {
		hash := sha512.New()
		hash.Write([]byte(t.Random.String()))
		sum := hash.Sum([]byte{})
		return security.Token{
			SHA512:   hex.EncodeToString(sum),
			OwnerUID: uuid.New().String(),
			IssuedAt: t.Random.Time().UTC(),
			Duration: time.Duration(t.Random.IntBetween(int(time.Second), int(time.Hour))),
		}
	})
	factory.RegisterType(release.Pilot{}, func() interface{} {
		return release.Pilot{
			FlagID:          ExampleReleaseFlag(t).ID,
			EnvironmentID:   ExampleDeploymentEnvironment(t).ID,
			PublicID:        uuid.New().String(),
			IsParticipating: t.Random.Bool(),
		}
	})
	factory.RegisterType(release.Rollout{}, func() interface{} {
		t.Helper()
		return release.Rollout{
			FlagID:                  ExampleReleaseFlag(t).ID,
			DeploymentEnvironmentID: ExampleDeploymentEnvironment(t).ID,
			Plan: func() release.RolloutDefinition {
				switch t.Random.IntN(3) {
				case 0:
					byPercentage := release.NewRolloutDecisionByPercentage()
					byPercentage.Percentage = t.Random.IntBetween(0, 100)
					return byPercentage

				case 1:
					byAPI := release.NewRolloutDecisionByAPI()
					u, err := url.ParseRequestURI(fmt.Sprintf(`https://example.com/%s`, url.PathEscape(t.Random.String())))
					if err != nil {
						panic(err.Error())
					}
					byAPI.URL = u
					return byAPI

				case 2:
					byPercentage := release.NewRolloutDecisionByPercentage()
					byPercentage.Percentage = t.Random.IntBetween(0, 100)

					byAPI := release.NewRolloutDecisionByAPI()
					u, err := url.ParseRequestURI(fmt.Sprintf(`https://example.com/%s`, url.PathEscape(t.Random.String())))
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
			}(),
		}
	})
	return factory
}
