package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/adamluzsi/toggler/extintf/httpintf/httputils"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/usecases"
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

	if handleError(w, pu.CreateFeatureFlag(context.TODO(), &flag), http.StatusInternalServerError) {
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

	if handleError(w, pu.CreateFeatureFlag(context.TODO(), ff), http.StatusInternalServerError) {
		return
	}

}
