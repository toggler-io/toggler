package httpapi

import (
	context2 "context"
	"encoding/json"
	"net/http"

	"github.com/toggler-io/toggler/extintf/httpintf/httputils"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/usecases"
)

func (sm *ServeMux) UpdateFeatureFlagJSON(w http.ResponseWriter, r *http.Request) {

	pu := r.Context().Value(`ProtectedUsecases`).(*usecases.ProtectedUsecases)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	var flag release.Flag

	if handleError(w, decoder.Decode(&flag), http.StatusBadRequest) {
		return
	}

	if handleError(w, pu.UpdateFeatureFlag(context2.TODO(), &flag), http.StatusInternalServerError) {
		return
	}

	serveJSON(w, map[string]interface{}{})

}

func (sm *ServeMux) UpdateFeatureFlagFORM(w http.ResponseWriter, r *http.Request) {

	pu := r.Context().Value(`ProtectedUsecases`).(*usecases.ProtectedUsecases)

	if handleError(w, r.ParseForm(), http.StatusBadRequest) {
		return
	}

	defer r.Body.Close() // ignorable

	ff, err := httputils.ParseFlagFromForm(r)

	if handleError(w, err, http.StatusBadRequest) {
		return
	}

	if ff.ID == `` {
		http.Error(w, `expected flag id not received`, http.StatusBadRequest)
		return
	}

	if ff.Name == `` {
		http.Error(w, `missing flag name`, http.StatusBadRequest)
		return
	}

	if handleError(w, pu.UpdateFeatureFlag(context2.TODO(), ff), http.StatusInternalServerError) {
		return
	}

}
