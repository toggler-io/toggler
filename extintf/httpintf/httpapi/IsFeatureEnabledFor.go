package httpapi

import (
	"encoding/json"
	"net/http"
)

// IsFeatureEnabledRequestParameters defines the parameters that
// swagger:parameters IsFeatureEnabled
type IsFeatureEnabledRequestParameters struct {
	// in: body
	Body IsFeatureEnabledRequestPayload
}

type IsFeatureEnabledRequestPayload struct {
	// Feature is the Feature Flag name that is needed to be checked for enrollment
	//
	// required: true
	// example: rollout-feature-flag
	Feature string `json:"feature"`
	// PilotID is the public unique ID of the pilot who's enrollment needs to be checked.
	//
	// required: true
	// example: pilot-public-id
	PilotID string `json:"id"`
}

type IsFeatureEnabledResponseBody = EnrollmentResponseBody

/*

	swagger:route POST /rollout/is-feature-enabled.json feature-flag pilot IsFeatureEnabled

	Check Rollout Feature Status For Pilot

	Reply back whether the feature for a given pilot id is enabled or not.
	By Default, this will be determined whether the flag exist,
	the pseudo random dice roll enrolls the pilot,
	or if there any manually set enrollment status for the pilot.
	The endpoint can be called with HTTP GET method as well,
	POST is used officially only to support most highly abstracted http clients.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Responses:
		  200: enrollmentResponse
		  400: errorResponse
		  500: errorResponse

*/
func (sm *ServeMux) IsFeatureEnabledFor(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var featureName, pilotID string

	q := r.URL.Query()
	featureName = q.Get(`feature`)
	pilotID = q.Get(`id`)

	if pilotID == `` || featureName == `` {
		if handleError(w, parseJSONPayloadForIsFeatureenabled(r, &featureName, &pilotID), http.StatusBadRequest) {
			return
		}
	}

	enrollment, err := sm.UseCases.IsFeatureEnabledFor(featureName, pilotID)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var resp IsFeatureEnabledResponseBody
	resp.Enrollment = enrollment

	serveJSON(w, &resp)

}

func parseJSONPayloadForIsFeatureenabled(r *http.Request, featureName, pilotID *string) error {
	jsondec := json.NewDecoder(r.Body)
	var payload IsFeatureEnabledRequestPayload
	if err := jsondec.Decode(&payload); err != nil {
		return err
	}

	*featureName = payload.Feature
	*pilotID = payload.PilotID
	return nil
}
