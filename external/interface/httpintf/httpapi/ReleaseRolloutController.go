package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/gorest"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
)

func NewReleaseRolloutHandler(uc *toggler.UseCases) http.Handler {
	c := ReleaseRolloutController{UseCases: uc}
	h := gorest.NewHandler(c)
	return httputils.AuthMiddleware(h, uc, ErrorWriterFunc)
}

type ReleaseRolloutController struct {
	UseCases *toggler.UseCases
}

//--------------------------------------------------------------------------------------------------------------------//

func (ctrl ReleaseRolloutController) getDeploymentEnvironment(ctx context.Context) release.Environment {
	return ctx.Value(DeploymentEnvironmentContextKey{}).(release.Environment)
}

func (ctrl ReleaseRolloutController) getReleaseFlag(ctx context.Context) release.Flag {
	return ctx.Value(ReleaseFlagContextKey{}).(release.Flag)
}

func (ctrl ReleaseRolloutController) handleFlagValidationError(w http.ResponseWriter, err error) bool {
	switch err {
	case release.ErrNameIsEmpty,
		release.ErrMissingFlag,
		release.ErrInvalidAction,
		release.ErrFlagAlreadyExist,
		release.ErrInvalidRequestURL,
		release.ErrInvalidPercentage:
		return handleError(w, err, http.StatusBadRequest)

	default:
		return handleError(w, err, http.StatusInternalServerError)
	}
}

//--------------------------------------------------------------------------------------------------------------------//

// CreateReleaseRolloutRequest
// swagger:parameters createReleaseRollout
type CreateReleaseRolloutRequest struct {
	// required: true
	// in: body
	Body struct {
		Rollout struct {
			EnvironmentID string `json:"env_id"`
			FlagID        string `json:"flag_id"`
			// Plan holds the composited rule set about the pilot participation decision logic.
			//
			// required: true
			// example: {"type": "percentage","percentage":42,"seed":10240}
			Plan interface{} `json:"plan"`
		} `json:"rollout"`
	}
}

// CreateReleaseRolloutResponse
// swagger:response createReleaseRolloutResponse
type CreateReleaseRolloutResponse struct {
	// in: body
	Body struct {
		Rollout Rollout `json:"rollout"`
	}
}

/*

	Create
	swagger:route POST /release-rollouts rollout createReleaseRollout

	This operation allows you to create a new release rollout.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: createReleaseRolloutResponse
		  400: errorResponse
		  401: errorResponse
		  500: errorResponse

*/
func (ctrl ReleaseRolloutController) Create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	type Payload struct {
		Rollout release.Rollout `json:"rollout"`
	}
	var p Payload

	if handleError(w, decoder.Decode(&p), http.StatusBadRequest) {
		return
	}

	rr := p.Rollout
	ctx := r.Context()
	rrs := ctrl.UseCases.Storage.ReleaseRollout(ctx)
	err := rrs.Create(ctx, &rr)

	if ctrl.handleFlagValidationError(w, err) {
		return
	}

	var resp CreateReleaseRolloutResponse
	resp.Body.Rollout.ID = p.Rollout.ID
	resp.Body.Rollout.FlagID = p.Rollout.FlagID
	resp.Body.Rollout.EnvironmentID = p.Rollout.DeploymentEnvironmentID
	resp.Body.Rollout.Plan = release.RolloutDefinitionView{Definition: p.Rollout.Plan}
	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

// ListReleaseRolloutResponse
// swagger:response listReleaseRolloutResponse
type ListReleaseRolloutResponse struct {
	// in: body
	Body struct {
		Rollouts []Rollout `json:"rollouts"`
	}
}

/*

	List
	swagger:route GET /release-rollouts rollout listReleaseRollouts

	List all the release flag that can be used to manage a feature rollout.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: listReleaseRolloutResponse
		  401: errorResponse
		  500: errorResponse

*/
func (ctrl ReleaseRolloutController) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var resp ListReleaseRolloutResponse

	err := iterators.ForEach(
		ctrl.UseCases.Storage.ReleaseRollout(ctx).FindAll(ctx),
		func(r release.Rollout) error {
			resp.Body.Rollouts = append(resp.Body.Rollouts, Rollout{
				ID:            r.ID,
				FlagID:        r.FlagID,
				EnvironmentID: r.DeploymentEnvironmentID,
				Plan:          release.RolloutDefinitionView{Definition: r.Plan},
			})

			return nil
		},
	)

	if handleError(w, err, http.StatusInternalServerError) {
		return
	}

	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

type ReleaseRolloutContextKey struct{}

func (ctrl ReleaseRolloutController) ContextWithResource(ctx context.Context, resourceID string) (context.Context, bool, error) {
	s := ctrl.UseCases.Storage.ReleaseRollout(ctx)
	var r release.Rollout
	found, err := s.FindByID(ctx, &r, resourceID)
	if err != nil {
		return ctx, false, err
	}
	if !found {
		return ctx, false, nil
	}
	return context.WithValue(ctx, ReleaseRolloutContextKey{}, r), true, nil
}

//--------------------------------------------------------------------------------------------------------------------//

// UpdateReleaseRolloutRequest
// swagger:parameters updateReleaseRollout
type UpdateReleaseRolloutRequest struct {
	// RolloutID is the rollout id
	//
	// in: path
	// required: true
	RolloutID string `json:"rolloutID"`
	// in: body
	Body struct {
		Rollout struct {
			Plan interface{} `json:"plan"`
		} `json:"rollout"`
	}
}

// UpdateReleaseRolloutResponse
// swagger:response updateReleaseRolloutResponse
type UpdateReleaseRolloutResponse struct {
	// in: body
	Body struct {
		Rollout struct {
			Plan interface{} `json:"plan"`
		} `json:"rollout"`
	}
}

/*

	Update
	swagger:route PUT /release-rollouts/{rolloutID} rollout updateReleaseRollout

	Update a release flag.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: updateReleaseRolloutResponse
		  400: errorResponse
		  500: errorResponse

*/
func (ctrl ReleaseRolloutController) Update(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	type Payload struct {
		Rollout struct {
			Plan release.RolloutDefinitionView `json:"plan"`
		} `json:"rollout"`
	}
	var p Payload

	if handleError(w, decoder.Decode(&p), http.StatusBadRequest) {
		return
	}

	ctx := r.Context()
	rollout := ctx.Value(ReleaseRolloutContextKey{}).(release.Rollout)
	rollout.Plan = p.Rollout.Plan.Definition

	if ctrl.handleFlagValidationError(w, ctrl.UseCases.Storage.ReleaseRollout(r.Context()).Update(r.Context(), &rollout)) {
		return
	}

	var resp UpdateReleaseRolloutResponse
	resp.Body.Rollout.Plan = release.RolloutDefinitionView{Definition: rollout.Plan}
	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

// DeleteReleaseRolloutRequest
// swagger:parameters deleteReleaseRollout
type DeleteReleaseRolloutRequest struct {
	// RolloutID is the rollout id
	//
	// in: path
	// required: true
	RolloutID string `json:"rolloutID"`
}

// DeleteReleaseRolloutResponse
// swagger:response deleteReleaseRolloutResponse
type DeleteReleaseRolloutResponse struct {
}

/*

	Delete
	swagger:route DELETE /release-rollouts/{rolloutID} rollout deleteReleaseRollout

	Delete a release rollout.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: deleteReleaseRolloutResponse
		  400: errorResponse
		  500: errorResponse

*/
func (ctrl ReleaseRolloutController) Delete(w http.ResponseWriter, r *http.Request) {
	ID := r.Context().Value(ReleaseRolloutContextKey{}).(release.Rollout).ID

	err := ctrl.UseCases.Storage.ReleaseRollout(r.Context()).DeleteByID(r.Context(), ID)
	if handleError(w, err, http.StatusBadRequest) {
		return
	}

	w.WriteHeader(200)
}

//--------------------------------------------------------------------------------------------------------------------//

type Rollout struct {
	ID            string      `json:"id"`
	FlagID        string      `json:"flag_id"`
	EnvironmentID string      `json:"env_id"`
	Plan          interface{} `json:"plan"`
}
