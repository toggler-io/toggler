package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
	"net/url"
	"strconv"
)

func (sm *ServeMux) CreateFeatureFlagJSON(w http.ResponseWriter, r *http.Request) {

	pu := r.Context().Value(`ProtectedUsecases`).(*usecases.ProtectedUsecases)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	var flag rollouts.FeatureFlag

	if handleError(w, decoder.Decode(&flag), http.StatusBadRequest) {
		return
	}

	if handleError(w, pu.CreateFeatureFlag(&flag), http.StatusInternalServerError) {
		return
	}

	serveJSON(w, 200, map[string]interface{}{})

}

func (sm *ServeMux) CreateFeatureFlagFORM(w http.ResponseWriter, r *http.Request) {

	pu := r.Context().Value(`ProtectedUsecases`).(*usecases.ProtectedUsecases)

	if handleError(w, r.ParseForm(), http.StatusBadRequest) {
		return
	}

	defer r.Body.Close() // ignorable

	var flag rollouts.FeatureFlag

	flag.Name = r.Form.Get(`flag.feature`)

	if flag.Name == `` {
		handleError(w, fmt.Errorf(`missing feature name`), http.StatusBadRequest)
		return
	}

	var randSeedSalt int64

	rawRandSeedSalt := r.Form.Get(`flag.rollout.randSeedSalt`)

	if rawRandSeedSalt != `` {
		var err error
		randSeedSalt, err = strconv.ParseInt(rawRandSeedSalt, 10, 64)

		if handleError(w, err, http.StatusBadRequest) {
			return
		}
	}

	flag.Rollout.RandSeedSalt = randSeedSalt

	percentage, err := strconv.ParseInt(r.Form.Get(`flag.rollout.strategy.percentage`), 10, 32)

	if handleError(w, err, http.StatusBadRequest) {
		return
	}

	flag.Rollout.Strategy.Percentage = int(percentage)

	var decisionLogicAPI *url.URL
	rawURL := r.Form.Get(`flag.rollout.strategy.decisionLogicApi`)

	if rawURL != `` {
		var err error
		decisionLogicAPI, err = url.ParseRequestURI(rawURL)

		if handleError(w, err, http.StatusBadRequest) {
			return
		}
	}

	flag.Rollout.Strategy.DecisionLogicAPI = decisionLogicAPI

	if handleError(w, pu.CreateFeatureFlag(&flag), http.StatusInternalServerError) {
		return
	}

}
