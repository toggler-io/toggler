// TODO this is a MVP, a POC for pilot resource management, not yet tested.
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/adamluzsi/gorest"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
)

func NewReleasePilotHandler(uc *toggler.UseCases) *gorest.Handler {
	c := ReleasePilotController{UseCases: uc}
	h := gorest.NewHandler(struct {
		gorest.ContextHandler
		gorest.CreateController
		gorest.ListController
		gorest.UpdateController
	}{
		ContextHandler:   c,
		CreateController: gorest.AsCreateController(httputils.AuthMiddleware(http.HandlerFunc(c.Create), uc, ErrorWriterFunc)),
		ListController:   gorest.AsListController(httputils.AuthMiddleware(http.HandlerFunc(c.List), uc, ErrorWriterFunc)),
		UpdateController: gorest.AsUpdateController(httputils.AuthMiddleware(http.HandlerFunc(c.Update), uc, ErrorWriterFunc)),
	})
	return h
}

type ReleasePilotController struct {
	UseCases *toggler.UseCases
}

var _ release.ManualPilot

type ReleasePilotView struct {
	// ID represent the fact that this object will be persistent in the Subject
	ID string `ext:"ID"`
	// ReleasePilotID is the reference ID that can tell where this user record belongs to.
	ReleasePilotID string `json:"release_flag_id"`
	// ExternalID is the unique id that connect links a pilot with the caller services.
	// The caller service is the service that use the release toggles for example and need A/B testing or Canary launch.
	ExternalID string `json:"external_id"`
	// IsParticipating states that whether the ManualPilot for the given feature is enrolled, or blacklisted
	Enrolled bool `json:"enrolled"`
}

func (ReleasePilotView) FromReleasePilot(pilot release.ManualPilot) ReleasePilotView {
	var v ReleasePilotView
	v.ID = pilot.ID
	v.ReleasePilotID = pilot.FlagID
	v.ExternalID = pilot.ExternalID
	v.Enrolled = pilot.IsParticipating
	return v
}

func (v ReleasePilotView) ToReleasePilot() release.ManualPilot {
	var pilot release.ManualPilot
	pilot.ID = v.ID
	pilot.FlagID = v.ReleasePilotID
	pilot.ExternalID = v.ExternalID
	pilot.IsParticipating = v.Enrolled
	return pilot
}

//--------------------------------------------------------------------------------------------------------------------//

//--------------------------------------------------------------------------------------------------------------------//

// CreateReleasePilotRequest
// swagger:parameters createReleasePilot
type CreateReleasePilotRequest struct {
	// ReleaseFlagID is the release flag id or the alias name.
	//
	// in: path
	// required: true
	ReleaseFlagID string `json:"flagID"`
	// in: body
	Body struct {
		Pilot ReleasePilotView `json:"pilot"`
	}
}

// CreateReleasePilotResponse
// swagger:response createReleasePilotResponse
type CreateReleasePilotResponse struct {
	// in: body
	Body struct {
		Pilot ReleasePilotView `json:"pilot"`
	}
}

/*

	Create
	swagger:route POST /release-flags/{flagID}/pilots release feature flag createReleasePilot

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
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	var req CreateReleasePilotRequest

	if handleError(w, decoder.Decode(&req.Body), http.StatusBadRequest) {
		return
	}

	req.Body.Pilot.ID = `` // ignore id if given
	pilot := req.Body.Pilot.ToReleasePilot()

	if err := ctrl.UseCases.SetPilotEnrollmentForFeature(r.Context(), pilot.FlagID, pilot.DeploymentEnvironmentID, pilot.ExternalID, pilot.IsParticipating); handleError(w, err, http.StatusInternalServerError) {
		return
	}

	p, err := ctrl.UseCases.RolloutManager.Storage.FindReleaseManualPilotByExternalID(r.Context(), pilot.FlagID, pilot.DeploymentEnvironmentID, pilot.ExternalID)
	if handleError(w, err, http.StatusInternalServerError) {
		return
	}
	if p == nil {
		p = &pilot
	}

	var resp CreateReleasePilotResponse
	resp.Body.Pilot = resp.Body.Pilot.FromReleasePilot(*p)
	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

// ListReleasePilotRequest
// swagger:parameters listReleasePilots
type ListReleasePilotRequest struct {
	// ReleaseFlagID is the release flag id or the alias name.
	//
	// in: path
	// required: true
	ReleaseFlagID string `json:"flagID"`
}

// ListReleasePilotResponse
// swagger:response listReleasePilotResponse
type ListReleasePilotResponse struct {
	// in: body
	Body struct {
		Pilots []ReleasePilotView `json:"pilots"`
	}
}

/*

	List
	swagger:route GET /release-flags/{flagID}/pilots release feature flag listReleasePilots

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
	flag := r.Context().Value(ReleaseFlagContextKey{}).(release.Flag)
	//FIXME
	pilotsIter := ctrl.UseCases.RolloutManager.Storage.FindReleasePilotsByReleaseFlag(r.Context(), flag)
	defer pilotsIter.Close()

	var resp ListReleasePilotResponse
	for pilotsIter.Next() {
		var p release.ManualPilot

		if handleError(w, pilotsIter.Decode(&p), http.StatusInternalServerError) {
			return
		}

		resp.Body.Pilots = append(resp.Body.Pilots, ReleasePilotView{}.FromReleasePilot(p))
	}

	if handleError(w, pilotsIter.Err(), http.StatusInternalServerError) {
		return
	}

	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

type ReleasePilotContextKey struct{}

func (ctrl ReleasePilotController) ContextWithResource(ctx context.Context, pilotID string) (context.Context, bool, error) {
	var p release.ManualPilot
	found, err := ctrl.UseCases.RolloutManager.Storage.FindByID(ctx, &p, pilotID)
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
	// ReleaseFlagID is the release flag id or the alias name.
	//
	// in: path
	// required: true
	ReleaseFlagID string `json:"flagID"`
	// PilotID is the release flag id or the alias name.
	//
	// in: path
	// required: true
	PilotID string `json:"pilotID"`
	// in: body
	Body struct {
		Pilot ReleasePilotView `json:"pilot"`
	}
}

// UpdateReleasePilotResponse
// swagger:response updateReleasePilotResponse
type UpdateReleasePilotResponse struct {
	// in: body
	Body struct {
		Pilot ReleasePilotView `json:"pilot"`
	}
}

/*

	Update
	swagger:route PUT /release-flags/{flagID}/pilots/{pilotID} release feature flag pilot updateReleasePilot

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
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	var req CreateReleasePilotRequest

	if handleError(w, decoder.Decode(&req.Body), http.StatusBadRequest) {
		return
	}

	req.Body.Pilot.ID = r.Context().Value(ReleasePilotContextKey{}).(release.ManualPilot).ID
	pilot := req.Body.Pilot.ToReleasePilot()

	if err := ctrl.UseCases.SetPilotEnrollmentForFeature(r.Context(), pilot.FlagID, pilot.DeploymentEnvironmentID, pilot.ExternalID, pilot.IsParticipating); handleError(w, err, http.StatusInternalServerError) {
		return
	}

	p, err := ctrl.UseCases.Storage.FindReleaseManualPilotByExternalID(r.Context(), pilot.FlagID, pilot.DeploymentEnvironmentID, pilot.ExternalID)
	if handleError(w, err, http.StatusInternalServerError) {
		return
	}
	if p == nil {
		p = &pilot
	}

	var resp CreateReleasePilotResponse
	resp.Body.Pilot = resp.Body.Pilot.FromReleasePilot(*p)
	serveJSON(w, resp.Body)
}
