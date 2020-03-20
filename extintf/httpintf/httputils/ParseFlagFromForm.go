package httputils

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/toggler-io/toggler/domains/release"
)

func ParseFlagFromForm(r *http.Request) (*release.Flag, error) {

	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	var flag release.Flag

	flag.Name = r.Form.Get(`flag.name`)
	flag.ID = r.Form.Get(`flag.id`)

	var randSeedSalt int64

	rawRandSeedSalt := r.Form.Get(`flag.rollout.randSeed`)

	if rawRandSeedSalt != `` {

		var err error
		randSeedSalt, err = strconv.ParseInt(rawRandSeedSalt, 10, 64)

		if err != nil {
			return nil, err
		}

	}

	flag.Rollout.RandSeed = randSeedSalt

	percentage, err := strconv.ParseInt(r.Form.Get(`flag.rollout.strategy.percentage`), 10, 32)

	if err != nil {
		return nil, err
	}

	flag.Rollout.Strategy.Percentage = int(percentage)

	var decisionLogicAPI *url.URL
	rawURL := r.Form.Get(`flag.rollout.strategy.decisionLogicApi`)

	if rawURL != `` {
		var err error
		decisionLogicAPI, err = url.ParseRequestURI(rawURL)

		if err != nil {
			return nil, err
		}
	}

	flag.Rollout.Strategy.DecisionLogicAPI = decisionLogicAPI

	return &flag, nil

}
