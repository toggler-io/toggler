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

func NewReleaseFlagHandler(uc *toggler.UseCases) *gorest.Handler {
	c := ReleaseFlagController{UseCases: uc}
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

type ReleaseFlagController struct {
	UseCases *toggler.UseCases
}

//--------------------------------------------------------------------------------------------------------------------//

func (ctrl ReleaseFlagController) handleFlagValidationError(w http.ResponseWriter, err error) bool {
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

// CreateReleaseFlagRequest
// swagger:parameters createReleaseFlag
type CreateReleaseFlagRequest struct {
	// in: body
	Body struct {
		Flag release.Flag `json:"flag"`
	}
}

// CreateReleaseFlagResponse
// swagger:response createReleaseFlagResponse
type CreateReleaseFlagResponse struct {
	// in: body
	Body struct {
		Flag release.Flag `json:"flag"`
	}
}

/*

	Create
	swagger:route POST /release-flags flag createReleaseFlag

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
		  200: createReleaseFlagResponse
		  400: errorResponse
		  401: errorResponse
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

	req.Body.Flag.ID = `` // ignore id if given
	flag := req.Body.Flag

	if ctrl.handleFlagValidationError(w, ctrl.UseCases.CreateFeatureFlag(r.Context(), &flag)) {
		return
	}

	var resp CreateReleaseFlagResponse
	resp.Body.Flag = flag
	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

// ListReleaseFlagRequest
// swagger:parameters listReleaseFlags
type ListReleaseFlagRequest struct {
}

// ListReleaseFlagResponse
// swagger:response listReleaseFlagResponse
type ListReleaseFlagResponse struct {
	// in: body
	Body struct {
		Flags []release.Flag `json:"flags"`
	}
}

/*

	List
	swagger:route GET /release-flags flag listReleaseFlags

	List all the release flag that can be used to manage a feature rollout.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: listReleaseFlagResponse
		  401: errorResponse
		  500: errorResponse

*/
func (ctrl ReleaseFlagController) List(w http.ResponseWriter, r *http.Request) {
	rfs, err := ctrl.UseCases.ListFeatureFlags(r.Context())

	if handleError(w, err, http.StatusInternalServerError) {
		return
	}

	var resp ListReleaseFlagResponse
	resp.Body.Flags = rfs
	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

type ReleaseFlagContextKey struct{}

func (ctrl ReleaseFlagController) ContextWithResource(ctx context.Context, resourceID string) (context.Context, bool, error) {
	s := ctrl.UseCases.Storage
	//flag, err := s.FindReleaseFlagByName(ctx, resourceID)
	//if err != nil {
	//	return ctx, false, err
	//}
	//if flag != nil {
	//	return context.WithValue(ctx, ReleaseFlagContextKey{}, *flag), true, nil
	//}

	var f release.Flag
	found, err := s.FindByID(ctx, &f, resourceID)
	if err != nil {
		return ctx, false, err
	}
	if !found {
		return ctx, false, nil
	}
	return context.WithValue(ctx, ReleaseFlagContextKey{}, f), true, nil
}

//--------------------------------------------------------------------------------------------------------------------//

// UpdateReleaseFlagRequest
// swagger:parameters updateReleaseFlag
type UpdateReleaseFlagRequest struct {
	// FlagID is the release flag id or the alias name.
	//
	// in: path
	// required: true
	FlagID string `json:"flagID"`
	// in: body
	Body struct {
		Flag release.Flag `json:"flag"`
	}
}

// UpdateReleaseFlagResponse
// swagger:response updateReleaseFlagResponse
type UpdateReleaseFlagResponse struct {
	// in: body
	Body struct {
		Flag release.Flag `json:"flag"`
	}
}

/*

	Update
	swagger:route PUT /release-flags/{flagID} flag updateReleaseFlag

	Update a release flag.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Security:
		  AppToken: []

		Responses:
		  200: updateReleaseFlagResponse
		  400: errorResponse
		  500: errorResponse

*/
func (ctrl ReleaseFlagController) Update(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	defer r.Body.Close() // ignorable

	var req UpdateReleaseFlagRequest

	if handleError(w, decoder.Decode(&req.Body), http.StatusBadRequest) {
		return
	}

	flag := req.Body.Flag
	flag.ID = r.Context().Value(ReleaseFlagContextKey{}).(release.Flag).ID

	if ctrl.handleFlagValidationError(w, ctrl.UseCases.UpdateFeatureFlag(r.Context(), &flag)) {
		return
	}

	var resp UpdateReleaseFlagResponse
	resp.Body.Flag = flag
	serveJSON(w, resp.Body)
}
