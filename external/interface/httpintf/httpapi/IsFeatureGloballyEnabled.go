package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
)

// IsFeatureGloballyEnabledRequestParameters is the expected payload
// that holds the feature flag name that needs to be observed from global rollout perspective.
// swagger:parameters IsFeatureGloballyEnabled
type IsFeatureGloballyEnabledRequestParameters struct {
	// in: body
	Body IsFeatureGloballyEnabledRequestBody
}

type IsFeatureGloballyEnabledRequestBody struct {
	// Feature is the Feature Flag name that is needed to be checked for enrollment
	//
	// required: true
	// example: rollout-feature-flag
	Feature string  `json:"feature"`
}

type IsFeatureGloballyEnabledResponseBody = EnrollmentResponseBody

/*

	swagger:route POST /release/is-feature-globally-enabled.json release-flag pilot IsFeatureGloballyEnabled

	Check Rollout Feature Status for Global use

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
		  200: enrollmentResponse
		  400: errorResponse
		  500: errorResponse

*/
func (sm *Handler) IsFeatureGloballyEnabled(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var featureName string

	q := r.URL.Query()
	featureName = q.Get(`feature`)

	if featureName == `` {
		if handleError(w, parseJSONPayloadForIsFeatureGloballyEnabled(r, &featureName), http.StatusBadRequest) {
			return
		}
	}

	enrollment, err := sm.UseCases.FlagChecker.IsFeatureGloballyEnabled(featureName)

	if err != nil {
		log.Println(`ERROR`, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var resp IsFeatureGloballyEnabledResponseBody
	resp.Enrollment = enrollment

	serveJSON(w, &resp)
}

func parseJSONPayloadForIsFeatureGloballyEnabled(r *http.Request, featureName *string) error {
	jsondec := json.NewDecoder(r.Body)
	var payload IsFeatureGloballyEnabledRequestBody
	if err := jsondec.Decode(&payload); err != nil {
		return err
	}

	*featureName = payload.Feature
	return nil
}
