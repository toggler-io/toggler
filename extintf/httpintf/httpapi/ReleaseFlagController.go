package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/usecases"
)

type ReleaseFlagController struct {
	UseCases *usecases.UseCases
}

// CreateReleaseFlagRequest
// swagger:parameters createReleaseFlag
type CreateReleaseFlagRequest struct {
	// in: body
	Body struct {
		Flag release.Flag `json:"flag"`
	}
}

// CreateReleaseFlagResponse returns
// swagger:response createReleaseFlagResponse
type CreateReleaseFlagResponse struct {
	// in: body
	Body struct {
		Flag release.Flag `json:"flag"`
	}
}

/*

	swagger:route POST /release-flags release feature flag createReleaseFlag

	Create a release flag that can be used for managing a feature rollout.
	This operation allows you to create a new release flag.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Responses:
		  200: createReleaseFlagResponse
		  400: errorResponse
		  500: errorResponse

*/
func (ctrl ReleaseFlagController) Create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	var req CreateReleaseFlagRequest

	if handleError(w, decoder.Decode(&req.Body), http.StatusBadRequest) {
		return
	}

	switch err := ctrl.UseCases.CreateFeatureFlag(r.Context(), &req.Body.Flag); err {
	case release.ErrNameIsEmpty,
		release.ErrMissingFlag,
		release.ErrInvalidAction,
		release.ErrFlagAlreadyExist,
		release.ErrInvalidRequestURL,
		release.ErrInvalidPercentage:
		if handleError(w, err, http.StatusBadRequest) {
			return
		}

	default:
		if handleError(w, err, http.StatusInternalServerError) {
			return
		}
	}

	var resp CreateReleaseFlagResponse
	resp.Body.Flag = req.Body.Flag
	serveJSON(w, resp.Body)
}
