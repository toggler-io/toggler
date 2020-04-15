package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/adamluzsi/gorest"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
	"github.com/toggler-io/toggler/usecases"
)

func NewReleaseFlagHandler(uc *usecases.UseCases) *gorest.Handler {
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
	h.Handle(`/global`, http.HandlerFunc(c.GetReleaseFlagGlobalStates))
	return h
}

type ReleaseFlagController struct {
	UseCases *usecases.UseCases
}

type ReleaseFlagView struct {
	// ID represent the fact that this object will be persistent in the Subject
	ID      string `ext:"ID" json:"id,omitempty"`
	Name    string `json:"name"`
	Rollout struct {
		// RandSeed allows you to configure the randomness for the percentage based pilot enrollment selection.
		// This value could have been neglected by using the flag name as random seed,
		// but that would reduce the flexibility for edge cases where you want
		// to use a similar pilot group as a successful flag rollout before.
		RandSeed int64 `json:"rand_seed_salt"`

		// Strategy expects to determines the behavior of the rollout workflow.
		// the actual behavior implementation is with the RolloutManager,
		// but the configuration data is located here
		Strategy struct {
			// Percentage allows you to define how many of your user base should be enrolled pseudo randomly.
			Percentage int `json:"percentage"`
			// DecisionLogicAPI allow you to do rollout based on custom domain needs such as target groups,
			// which decision logic is available trough an API endpoint call
			DecisionLogicAPI *url.URL `json:"decision_logic_api"`
		} `json:"strategy"`
	} `json:"rollout"`
}

func (ReleaseFlagView) FromReleaseFlag(flag release.Flag) ReleaseFlagView {
	var v ReleaseFlagView
	v.ID = flag.ID
	v.Name = flag.Name
	v.Rollout.RandSeed = flag.Rollout.RandSeed
	v.Rollout.Strategy.DecisionLogicAPI = flag.Rollout.Strategy.DecisionLogicAPI
	v.Rollout.Strategy.Percentage = flag.Rollout.Strategy.Percentage
	return v
}

func (v ReleaseFlagView) ToReleaseFlag() release.Flag {
	var flag release.Flag
	flag.ID = v.ID
	flag.Name = v.Name
	flag.Rollout.RandSeed = v.Rollout.RandSeed
	flag.Rollout.Strategy.DecisionLogicAPI = v.Rollout.Strategy.DecisionLogicAPI
	flag.Rollout.Strategy.Percentage = v.Rollout.Strategy.Percentage
	return flag
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
		Flag ReleaseFlagView `json:"flag"`
	}
}

// CreateReleaseFlagResponse
// swagger:response createReleaseFlagResponse
type CreateReleaseFlagResponse struct {
	// in: body
	Body struct {
		Flag ReleaseFlagView `json:"flag"`
	}
}

/*

	Create
	swagger:route POST /release-flags release feature flag createReleaseFlag

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
	flag := req.Body.Flag.ToReleaseFlag()

	if ctrl.handleFlagValidationError(w, ctrl.UseCases.CreateFeatureFlag(r.Context(), &flag)) {
		return
	}

	var resp CreateReleaseFlagResponse
	resp.Body.Flag = resp.Body.Flag.FromReleaseFlag(flag)
	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

// ListReleaseFlagRequest
// swagger:parameters listReleaseFlags
type ListReleaseFlagRequest struct {
	// in: body
	Body struct{}
}

// ListReleaseFlagResponse
// swagger:response listReleaseFlagResponse
type ListReleaseFlagResponse struct {
	// in: body
	Body struct {
		Flags []ReleaseFlagView `json:"flags"`
	}
}

/*

	List
	swagger:route GET /release-flags release feature flag listReleaseFlags

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
	for _, rf := range rfs {
		resp.Body.Flags = append(resp.Body.Flags, ReleaseFlagView{}.FromReleaseFlag(rf))
	}

	serveJSON(w, resp.Body)
}

//--------------------------------------------------------------------------------------------------------------------//

type ReleaseFlagContextKey struct{}

func (ctrl ReleaseFlagController) ContextWithResource(ctx context.Context, resourceID string) (context.Context, bool, error) {
	s := ctrl.UseCases.FlagChecker.Storage
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

// GetReleaseFlagGlobalStatesRequest is the expected payload
// that holds the feature flag name that needs to be observed from global rollout perspective.
// swagger:parameters getReleaseFlagGlobalStates
type GetReleaseFlagGlobalStatesRequest struct {
	// FlagID is the release flag id or the alias name.
	//
	// in: path
	// required: true
	FlagID string `json:"flagID"`
}

// GetReleaseFlagGlobalStatesResponse
// swagger:response getReleaseFlagGlobalStatesResponse
type GetReleaseFlagGlobalStatesResponse struct {
	// in: body
	Body struct {
		// Enrollment is the release feature flag enrollment status.
		Enrollment bool `json:"enrollment"`
	}
}

/*

	GetReleaseFlagGlobalStates
	swagger:route GET /release-flags/{flagID}/global release feature flag pilot getReleaseFlagGlobalStates

	Get Release flag statistics regarding global state by the name of the release flag.

	Reply back whether the feature rolled out globally or not.
	This is especially useful for cases where you don't have pilot id.
	Such case is batch processing, or dark launch flips.
	By Default, this will be determined whether the flag exist,
	Then  whether the release id done to everyone or not by percentage.
	The endpoint can be called with HTTP GET method as well,
	POST is used officially only to support most highly abstracted http clients.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Responses:
		  200: getReleaseFlagGlobalStatesResponse
		  400: errorResponse
		  500: errorResponse

*/
func (ctrl ReleaseFlagController) GetReleaseFlagGlobalStates(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	flag := r.Context().Value(ReleaseFlagContextKey{}).(release.Flag)
	enrollment := flag.Rollout.Strategy.IsGlobal()

	var resp GetReleaseFlagGlobalStatesResponse
	resp.Body.Enrollment = enrollment
	serveJSON(w, &resp.Body)
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
		Flag ReleaseFlagView `json:"flag"`
	}
}

// UpdateReleaseFlagResponse
// swagger:response updateReleaseFlagResponse
type UpdateReleaseFlagResponse struct {
	// in: body
	Body struct {
		Flag ReleaseFlagView `json:"flag"`
	}
}

/*

	Update
	swagger:route PUT /release-flags/{flagID} release feature flag pilot updateReleaseFlag

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

	flag := req.Body.Flag.ToReleaseFlag()
	flag.ID = r.Context().Value(ReleaseFlagContextKey{}).(release.Flag).ID

	if ctrl.handleFlagValidationError(w, ctrl.UseCases.UpdateFeatureFlag(r.Context(), &flag)) {
		return
	}

	var resp UpdateReleaseFlagResponse
	resp.Body.Flag = ReleaseFlagView{}.FromReleaseFlag(flag)
	serveJSON(w, resp.Body)
}
