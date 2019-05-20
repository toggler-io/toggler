package api

import (
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
	"strconv"
)

func (sm *ServeMux) UpdateFeatureFlagRolloutPercentage(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get(`token`)
	featureFlagName := r.URL.Query().Get(`feature`)
	percentageStr := r.URL.Query().Get(`percentage`)

	percentage, err := strconv.Atoi(percentageStr)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = sm.UseCases.UpdateFeatureFlagRolloutPercentage(token, featureFlagName, percentage)

	if err == usecases.ErrInvalidToken {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	serveJSON(w, 200, map[string]interface{}{})
}
