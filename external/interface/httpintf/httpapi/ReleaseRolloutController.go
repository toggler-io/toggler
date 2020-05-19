package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/gorest"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
	"github.com/toggler-io/toggler/usecases"
)

func NewReleaseRolloutHandler(uc *usecases.UseCases) *gorest.Handler {
	c := ReleaseRolloutController{UseCases: uc}
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
	//return httputils.AuthMiddleware(h, uc, ErrorWriterFunc)
}

type ReleaseRolloutController struct {
	UseCases *usecases.UseCases
}

//--------------------------------------------------------------------------------------------------------------------//

func (ctrl ReleaseRolloutController) getDeploymentEnvironment(ctx context.Context) deployment.Environment {
	return ctx.Value(DeploymentEnvironmentContextKey{}).(deployment.Environment)
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
	// EnvironmentID is the deployment environment ID
	//
	// in: path
	// required: true
	EnvironmentID string `json:"envID"`
	// FlagID is the release flag id
	//
	// in: path
	// required: true
	FlagID string `json:"flagID"`
	// in: body
	Body struct {
		Rollout struct {
			// Plan holds the composited rule set about the pilot participation decision logic.
			//
			// required: true
			// example: {"percentage":{"percentage":42,"seed":10240}}
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
	swagger:route POST /deployment-environments/{envID}/release-flags/{flagID}/release-rollouts release rollout flag createReleaseRollout

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

	ctx := r.Context()

	p.Rollout.FlagID = ctrl.getReleaseFlag(ctx).ID
	p.Rollout.DeploymentEnvironmentID = ctrl.getDeploymentEnvironment(ctx).ID

	if ctrl.handleFlagValidationError(w, ctrl.UseCases.Storage.Create(r.Context(), &p.Rollout)) {
		return
	}

	var resp CreateReleaseRolloutResponse
	resp.Body.Rollout.FlagID = p.Rollout.FlagID
	resp.Body.Rollout.EnvironmentID = p.Rollout.DeploymentEnvironmentID
	resp.Body.Rollout.Plan = release.RolloutDefinitionView{Definition: p.Rollout.Plan}
	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

// ListReleaseRolloutRequest
// swagger:parameters listReleaseRollouts
type ListReleaseRolloutRequest struct {
	// EnvironmentID is the deployment environment ID
	//
	// in: path
	// required: true
	EnvironmentID string `json:"envID"`
	// FlagID is the release flag id
	//
	// in: path
	// required: true
	FlagID string `json:"flagID"`
	// in: body
	Body struct{}
}

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
	swagger:route GET /deployment-environments/{envID}/release-flags/{flagID}/release-rollouts release rollout flag listReleaseRollouts

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
	env := ctrl.getDeploymentEnvironment(ctx)

	var resp ListReleaseRolloutResponse

	// TODO: replace with Storage Contract
	// TODO:DEBT
	err := iterators.ForEach(
		iterators.Filter(
			ctrl.UseCases.Storage.FindAll(ctx, release.Rollout{}),
			func(r release.Rollout) bool { return r.DeploymentEnvironmentID == env.ID },
		),
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
	s := ctrl.UseCases.Storage
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
	// EnvironmentID is the deployment environment ID
	//
	// in: path
	// required: true
	EnvironmentID string `json:"envID"`
	// FlagID is the release flag id
	//
	// in: path
	// required: true
	FlagID string `json:"flagID"`
	// FlagID is the release flag id or the alias name.
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
	swagger:route PUT /deployment-environments/{envID}/release-flags/{flagID}/release-rollouts/{rolloutID} release rollout flag pilot updateReleaseRollout

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

	if ctrl.handleFlagValidationError(w, ctrl.UseCases.Storage.Update(r.Context(), &rollout)) {
		return
	}

	var resp UpdateReleaseRolloutResponse
	resp.Body.Rollout.Plan = release.RolloutDefinitionView{Definition: rollout.Plan}
	serveJSON(w, resp.Body)
}

type Rollout struct {
	ID            string      `json:"id"`
	FlagID        string      `json:"flag_id"`
	EnvironmentID string      `json:"env_id"`
	Plan          interface{} `json:"plan"`
}
