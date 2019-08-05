package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/adamluzsi/toggler/extintf/httpintf/httputils"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/usecases"
)

// CreateRolloutFeatureFlagJSONParameters is the request object for creating feature flags.
// swagger:parameters CreateRolloutFeatureFlag
type CreateRolloutFeatureFlagJSONParameters struct {
	// in: body
	Body rollouts.FeatureFlag
}

// CreateRolloutFeatureFlagResponse returns information about the requester's rollout feature enrollment status.
// swagger:response createRolloutFeatureFlagResponse
type CreateRolloutFeatureFlagResponse struct {
	// in: body
	Body struct{}
}

/*

	swagger:route POST /rollout/flag/create.json rollout feature-flag CreateRolloutFeatureFlag

	Create Rollout Feature Flag

	This operation allows you to create a new rollout feature flag.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Responses:
		  200: createRolloutFeatureFlagResponse
		  400: errorResponse
		  500: errorResponse

*/
func (sm *ServeMux) CreateRolloutFeatureFlagJSON(w http.ResponseWriter, r *http.Request) {

	pu := r.Context().Value(`ProtectedUsecases`).(*usecases.ProtectedUsecases)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	var flag rollouts.FeatureFlag

	if handleError(w, decoder.Decode(&flag), http.StatusBadRequest) {
		return
	}

	if handleError(w, pu.CreateFeatureFlag(r.Context(), &flag), http.StatusInternalServerError) {
		return
	}

	serveJSON(w, struct{}{})

}

func (sm *ServeMux) CreateRolloutFeatureFlagFORM(w http.ResponseWriter, r *http.Request) {

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

	if handleError(w, pu.CreateFeatureFlag(r.Context(), ff), http.StatusInternalServerError) {
		return
	}

}
