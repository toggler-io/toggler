package httpapi

import (
	"encoding/json"
	"net/http"
)

type IsFeatureEnabledForReqBody struct {
	Feature string `json:"feature"`
	PilotID string `json:"id"`
}

/*
	swagger:response enrollmentResponse

		  content:
		    application/json:
		      examples:
		        '0':
		          value: |
		            {"enrollment":false}

 */

type EnrollmentResponse struct {
	// The status that tells the calller's enrollment state in the feature if the enrollment.
	// in: body
	// example: true
	Enrollment bool `json:"enrollment"`
}
/*

	swagger:route GET /api/v1/rollout/is-enabled.json feature-flag pilot IsFeatureEnabled

	Check Rollout Feature Status For Pilot

	Reply back whether the feature for a given pilot id is enabled or not.
	By Default, this will be determined whether the flag exist,
	the pseudo random dice roll enrolls the pilot,
	or if there any manually set enrollment status for the pilot.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: http, https

		Responses:
		  200: enrollmentResponse
		  400: errorResponseBody
		  500: errorResponseBody

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

	var resp struct {
		Enrollment bool `json:"enrollment"`
	}
	resp.Enrollment = enrollment

	serveJSON(w, &resp)

}

func parseJSONPayloadForIsFeatureenabled(r *http.Request, featureName, pilotID *string) error {
	jsondec := json.NewDecoder(r.Body)
	var payload IsFeatureEnabledForReqBody
	if err := jsondec.Decode(&payload); err != nil {
		return err
	}

	*featureName = payload.Feature
	*pilotID = payload.PilotID
	return nil
}
