package httpapi

import (
	"encoding/json"
	"net/http"
)

// RolloutClientConfigParameters defines the parameters that
// swagger:parameters RolloutClientConfig
type RolloutClientConfigParameters struct {
	// in: body
	Body RolloutClientConfigRequestBody
}

type RolloutClientConfigRequestBody struct {
	// PilotID is the public uniq id that identify the caller pilot
	//
	// required: true
	// example: public-uniq-pilot-id
	PilotID string `json:"id"`
	// Features are the list of flag name that should be matched against the pilot and state the enrollment for each.
	//
	// required: true
	// example: ["my-feature-flag"]
	Features []string `json:"features"`
}

// RolloutClientConfigResponse returns information about the requester's rollout feature enrollment statuses.
// swagger:response rolloutClientConfigResponse
type RolloutClientConfigResponse struct {
	// in: body
	Body RolloutClientConfigResponseBody
}

// RolloutClientConfigResponseBody will contain the requested feature flag states for a certain pilot.
// The content expected to be cached in some form of state container.
type RolloutClientConfigResponseBody struct {
	// States holds the requested rollout feature flag enrollment statuses.
	States map[string]bool `json:"states"`
}

/*

	swagger:route POST /rollout/config.json feature-flag pilot RolloutClientConfig

	Check Multiple Rollout Feature Status For A Certain Pilot

	Return all the flag states that was requested in the favor of a Pilot.
	This endpoint especially useful for Mobile & SPA apps.
	The endpoint can be called with HTTP GET method as well,
	POST is used officially only to support most highly abstracted http clients.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Responses:
		  200: rolloutClientConfigResponse
		  400: errorResponse
		  500: errorResponse

*/
func (sm *ServeMux) RolloutConfigJSON(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	payloadDecoder := json.NewDecoder(r.Body)

	var requestData RolloutClientConfigRequestBody

	parseErr := payloadDecoder.Decode(&requestData)

	if parseErr != nil {
		parseErr = nil
		q := r.URL.Query()
		requestData.PilotID = q.Get(`id`)
		requestData.Features = append([]string{}, q[`feature`]...)
		requestData.Features = append(requestData.Features, q[`feature[]`]...)
	}

	states, err := sm.UseCases.GetPilotFlagStates(r.Context(), requestData.PilotID, requestData.Features...)

	if handleError(w, err, http.StatusInternalServerError) {
		return
	}

	serveJSON(w, RolloutClientConfigResponseBody{States: states})
}
