package httpapi

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/usecases"
)

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

	switch err := ctrl.UseCases.CreateFeatureFlag(r.Context(), &flag); err {
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
		  400: errorResponse
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
