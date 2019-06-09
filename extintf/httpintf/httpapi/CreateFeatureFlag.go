package httpapi

import (
	"encoding/json"
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httputils"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
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

	ff, err := httputils.ParseFlagFromForm(r)

	if handleError(w, err, http.StatusBadRequest) {
		return
	}

	if ff.Name == `` {
		http.Error(w, `missing flag name`, http.StatusBadRequest)
		return
	}

	if ff.ID != `` {
		http.Error(w, `unexpected flag id received`, http.StatusBadRequest)
		return
	}

	if handleError(w, pu.CreateFeatureFlag(ff), http.StatusInternalServerError) {
		return
	}

}
