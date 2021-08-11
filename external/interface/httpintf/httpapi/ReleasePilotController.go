// TODO this is a MVP, a POC for pilot resource management, not yet tested.
package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adamluzsi/gorest"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
)

func NewReleasePilotHandler(uc *toggler.UseCases) http.Handler {
	c := ReleasePilotController{UseCases: uc}
	h := gorest.NewHandler(c)
	httputils.AuthMiddleware(h, uc, ErrorWriterFunc)
	return h
}

type ReleasePilotController struct {
	UseCases *toggler.UseCases
}

//--------------------------------------------------------------------------------------------------------------------//

// CreateReleasePilotRequest
// swagger:parameters createReleasePilot
type CreateReleasePilotRequest struct {
	// in: body
	Body struct {
		Pilot release.Pilot `json:"pilot"`
	}
}

// CreateReleasePilotResponse
// swagger:response createReleasePilotResponse
type CreateReleasePilotResponse struct {
	// in: body
	Body struct {
		Pilot release.Pilot `json:"pilot"`
	}
}

/*

	Create
	swagger:route POST /release-pilots pilot createReleasePilot

	Create a release flag that can be used for managing a feature rollout.
	This operation allows you to create a new release flag.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: createReleasePilotResponse
		  400: errorResponse
		  401: errorResponse
		  500: errorResponse

*/
func (ctrl ReleasePilotController) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close()

	var req CreateReleasePilotRequest

	if handleError(w, decoder.Decode(&req.Body), http.StatusBadRequest) {
		return
	}

	req.Body.Pilot.ID = `` // ignore id if given
	pilot := req.Body.Pilot

	if ctrl.validatePilot(w, pilot) {
		return
	}

	rps := ctrl.UseCases.Storage.ReleasePilot(ctx)

	if handleError(w, rps.Create(ctx, &pilot), http.StatusBadRequest) {
		return
	}

	var resp CreateReleasePilotResponse
	resp.Body.Pilot = pilot
	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

// ListReleasePilotResponse
// swagger:response listReleasePilotResponse
type ListReleasePilotResponse struct {
	// in: body
	Body struct {
		Pilots []release.Pilot `json:"pilots"`
	}
}

/*

	List
	swagger:route GET /release-pilots pilot listReleasePilots

	List all the release flag that can be used to manage a feature rollout.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: listReleasePilotResponse
		  401: errorResponse
		  500: errorResponse

*/
func (ctrl ReleasePilotController) List(w http.ResponseWriter, r *http.Request) {
	pilotsIter := ctrl.UseCases.RolloutManager.Storage.ReleasePilot(r.Context()).FindAll(r.Context())
	defer pilotsIter.Close()

	var resp ListReleasePilotResponse
	for pilotsIter.Next() {
		var p release.Pilot

		if handleError(w, pilotsIter.Decode(&p), http.StatusInternalServerError) {
			return
		}

		resp.Body.Pilots = append(resp.Body.Pilots, p)
	}

	if handleError(w, pilotsIter.Err(), http.StatusInternalServerError) {
		return
	}

	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

type ReleasePilotContextKey struct{}

func (ctrl ReleasePilotController) ContextWithResource(ctx context.Context, pilotID string) (context.Context, bool, error) {
	var p release.Pilot
	found, err := ctrl.UseCases.RolloutManager.Storage.ReleasePilot(ctx).FindByID(ctx, &p, pilotID)
	if err != nil {
		return ctx, false, err
	}
	if !found {
		return ctx, false, nil
	}
	return context.WithValue(ctx, ReleasePilotContextKey{}, p), true, nil
}

//--------------------------------------------------------------------------------------------------------------------//

// UpdateReleasePilotRequest
// swagger:parameters updateReleasePilot
type UpdateReleasePilotRequest struct {
	// PilotID is the pilot id.
	//
	// in: path
	// required: true
	PilotID string `json:"pilotID"`
	// in: body
	Body struct {
		Pilot release.Pilot `json:"pilot"`
	}
}

// UpdateReleasePilotResponse
// swagger:response updateReleasePilotResponse
type UpdateReleasePilotResponse struct {
	// in: body
	Body struct {
		Pilot release.Pilot `json:"pilot"`
	}
}

/*

	Update
	swagger:route PUT /release-pilots/{pilotID} pilot updateReleasePilot

	Update a release flag.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: updateReleasePilotResponse
		  400: errorResponse
		  500: errorResponse

*/
func (ctrl ReleasePilotController) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	var req CreateReleasePilotRequest

	if handleError(w, decoder.Decode(&req.Body), http.StatusBadRequest) {
		return
	}

	req.Body.Pilot.ID = ctx.Value(ReleasePilotContextKey{}).(release.Pilot).ID
	pilot := req.Body.Pilot

	if ctrl.validatePilot(w, pilot) {
		return
	}

	rps := ctrl.UseCases.Storage.ReleasePilot(ctx)

	if handleError(w, rps.Update(ctx, &pilot), http.StatusBadRequest) {
		return
	}

	var resp CreateReleasePilotResponse
	resp.Body.Pilot = pilot
	serveJSON(w, resp.Body)
}

func (ctrl ReleasePilotController) validatePilot(w http.ResponseWriter, pilot release.Pilot) bool {
	if pilot.FlagID == "" {
		handleError(w, fmt.Errorf("missing flag_id"), http.StatusBadRequest)
		return true
	}
	if pilot.EnvironmentID == "" {
		handleError(w, fmt.Errorf("missing env_id"), http.StatusBadRequest)
		return true
	}
	return false
}

//--------------------------------------------------------------------------------------------------------------------//

// DeleteReleasePilotRequest
// swagger:parameters deleteReleasePilot
type DeleteReleasePilotRequest struct {
	// PilotID is the pilot id.
	//
	// in: path
	// required: true
	PilotID string `json:"pilotID"`
}

// DeleteReleasePilotResponse
// swagger:response deleteReleasePilotResponse
type DeleteReleasePilotResponse struct {
}

/*

	Delete
	swagger:route DELETE /release-pilots/{pilotID} pilot deleteReleasePilot

	Delete a release pilot.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: deleteReleasePilotResponse
		  400: errorResponse
		  500: errorResponse

*/
func (ctrl ReleasePilotController) Delete(w http.ResponseWriter, r *http.Request) {
	ID := r.Context().Value(ReleasePilotContextKey{}).(release.Pilot).ID

	err := ctrl.UseCases.Storage.ReleasePilot(r.Context()).DeleteByID(r.Context(), ID)
	if handleError(w, err, http.StatusBadRequest) {
		return
	}

	w.WriteHeader(200)
}
