package httpapi

import (
	"context"
	"encoding/json"
	"github.com/toggler-io/toggler/domains/release"
	"net/http"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/gorest"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
)

func NewDeploymentEnvironmentHandler(uc *toggler.UseCases) http.Handler {
	c := DeploymentEnvironmentController{UseCases: uc}
	h := gorest.NewHandler(c)
	return httputils.AuthMiddleware(h, uc, ErrorWriterFunc)
}

type DeploymentEnvironmentController struct {
	UseCases *toggler.UseCases
}

//--------------------------------------------------------------------------------------------------------------------//

func (ctrl DeploymentEnvironmentController) handleValidationError(w http.ResponseWriter, err error) bool {
	switch err {
	case release.ErrEnvironmentNameIsEmpty:
		return handleError(w, err, http.StatusBadRequest)

	default:
		return handleError(w, err, http.StatusInternalServerError)
	}
}

//--------------------------------------------------------------------------------------------------------------------//

// CreateDeploymentEnvironmentRequest
// swagger:parameters createDeploymentEnvironment
type CreateDeploymentEnvironmentRequest struct {
	// in: body
	Body struct {
		Environment release.Environment `json:"environment"`
	}
}

// CreateDeploymentEnvironmentResponse
// swagger:response createDeploymentEnvironmentResponse
type CreateDeploymentEnvironmentResponse struct {
	// in: body
	Body struct {
		Environment release.Environment `json:"environment"`
	}
}

/*

	Create
	swagger:route POST /deployment-environments deployment createDeploymentEnvironment

	Create a deployment environment that can be used for managing a feature rollout.
	This operation allows you to create a new deployment environment.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: createDeploymentEnvironmentResponse
		  400: errorResponse
		  401: errorResponse
		  500: errorResponse

*/
func (ctrl DeploymentEnvironmentController) Create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	var req CreateDeploymentEnvironmentRequest

	if handleError(w, decoder.Decode(&req.Body), http.StatusBadRequest) {
		return
	}

	req.Body.Environment.ID = `` // ignore id if given
	env := req.Body.Environment

	if ctrl.handleValidationError(w, env.Validate()) {
		return
	}

	if ctrl.handleValidationError(w, ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).Create(r.Context(), &env)) {
		return
	}

	var resp CreateDeploymentEnvironmentResponse
	resp.Body.Environment = env
	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

// ListDeploymentEnvironmentRequest
// swagger:parameters listDeploymentEnvironments
type ListDeploymentEnvironmentRequest struct {
}

// ListDeploymentEnvironmentResponse
// swagger:response listDeploymentEnvironmentResponse
type ListDeploymentEnvironmentResponse struct {
	// in: body
	Body struct {
		Environments []release.Environment `json:"environments"`
	}
}

/*

	List
	swagger:route GET /deployment-environments deployment listDeploymentEnvironments

	List all the deployment environment that can be used to manage a feature rollout.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: listDeploymentEnvironmentResponse
		  401: errorResponse
		  500: errorResponse

*/
func (ctrl DeploymentEnvironmentController) List(w http.ResponseWriter, r *http.Request) {
	var resp ListDeploymentEnvironmentResponse

	if handleError(w,
		iterators.Collect(ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).FindAll(r.Context()), &resp.Body.Environments),
		http.StatusInternalServerError,
	) {
		return
	}

	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

type DeploymentEnvironmentContextKey struct{}

func (ctrl DeploymentEnvironmentController) ContextWithResource(ctx context.Context, resourceID string) (context.Context, bool, error) {
	s := ctrl.UseCases.Storage.ReleaseEnvironment(ctx)

	var f release.Environment
	found, err := s.FindByID(ctx, &f, resourceID)
	if err != nil {
		return ctx, false, err
	}
	if !found {
		return ctx, false, nil
	}
	return context.WithValue(ctx, DeploymentEnvironmentContextKey{}, f), true, nil
}

//--------------------------------------------------------------------------------------------------------------------//

// UpdateDeploymentEnvironmentRequest
// swagger:parameters updateDeploymentEnvironment
type UpdateDeploymentEnvironmentRequest struct {
	// EnvironmentID is the deployment environment id or the alias name.
	//
	// in: path
	// required: true
	EnvironmentID string `json:"envID"`
	// in: body
	Body struct {
		Environment release.Environment `json:"environment"`
	}
}

// UpdateDeploymentEnvironmentResponse
// swagger:response updateDeploymentEnvironmentResponse
type UpdateDeploymentEnvironmentResponse struct {
	// in: body
	Body struct {
		Environment release.Environment `json:"environment"`
	}
}

/*

	Update
	swagger:route PUT /deployment-environments/{envID} deployment updateDeploymentEnvironment

	Update a deployment environment.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: updateDeploymentEnvironmentResponse
		  400: errorResponse
		  500: errorResponse

*/
func (ctrl DeploymentEnvironmentController) Update(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	var req UpdateDeploymentEnvironmentRequest

	if handleError(w, decoder.Decode(&req.Body), http.StatusBadRequest) {
		return
	}

	env := req.Body.Environment
	env.ID = r.Context().Value(DeploymentEnvironmentContextKey{}).(release.Environment).ID

	if ctrl.handleValidationError(w, env.Validate()) {
		return
	}

	if ctrl.handleValidationError(w, ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).Update(r.Context(), &env)) {
		return
	}

	var resp UpdateDeploymentEnvironmentResponse
	resp.Body.Environment = env
	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

// DeleteDeploymentEnvironmentRequest
// swagger:parameters deleteDeploymentEnvironment
type DeleteDeploymentEnvironmentRequest struct {
	// EnvironmentID is the deployment environment id or the alias name.
	//
	// in: path
	// required: true
	EnvironmentID string `json:"envID"`
}

// DeleteDeploymentEnvironmentResponse
// swagger:response deleteDeploymentEnvironmentResponse
type DeleteDeploymentEnvironmentResponse struct {
}

/*

	Delete
	swagger:route DELETE /deployment-environments/{envID} deployment deleteDeploymentEnvironment

	Delete a deployment environment.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: deleteDeploymentEnvironmentResponse
		  400: errorResponse
		  500: errorResponse

*/

func (ctrl DeploymentEnvironmentController) Delete(w http.ResponseWriter, r *http.Request) {
	ID := r.Context().Value(DeploymentEnvironmentContextKey{}).(release.Environment).ID

	err := ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).DeleteByID(r.Context(), ID)
	if handleError(w, err, http.StatusBadRequest) {
		return
	}

	w.WriteHeader(200)
}
