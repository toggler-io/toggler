package httpapi

import (
	"encoding/json"
	"net/http"
)

// IsFeatureGloballyEnabledRequestParameters is the expected payload
// that holds the feature flag name that needs to be observed from global rollout perspective.
// swagger:parameters IsFeatureGloballyEnabled
type IsFeatureGloballyEnabledRequestParameters struct {
	// in: body
	Body IsFeatureGloballyEnabledRequestPayload
}

type IsFeatureGloballyEnabledRequestPayload struct {
	// Feature is the Feature Flag name that is needed to be checked for enrollment
	//
	// required: true
	// example: rollout-feature-flag
	Feature string  `json:"feature"`
}

type IsFeatureGloballyEnabledResponseBody = IsFeatureEnabledRequestPayload

/*

	swagger:route GET /api/v1/rollout/is-globally-enabled.json feature-flag pilot IsFeatureGloballyEnabled

	Check Rollout Feature Status for Global use

	Reply back whether the feature rolled out globally or not.
	This is especially useful for cases where you don't have pilot id.
	Such case is batch processing, or dark launch flips.
	By Default, this will be determined whether the flag exist,
	Then  whether the release id done to everyone or not by percentage.

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
func (sm *ServeMux) IsFeatureGloballyEnabled(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var featureName string

	q := r.URL.Query()
	featureName = q.Get(`feature`)

	if featureName == `` {
		if handleError(w, parseJSONPayloadForIsFeatureGloballyEnabled(r, &featureName), http.StatusBadRequest) {
			return
		}
	}

	enrollment, err := sm.UseCases.IsFeatureGloballyEnabled(featureName)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var resp struct{ Enrollment bool `json:"enrollment"` }
	resp.Enrollment = enrollment

	serveJSON(w, &resp)
}

func parseJSONPayloadForIsFeatureGloballyEnabled(r *http.Request, featureName *string) error {
	jsondec := json.NewDecoder(r.Body)
	var payload IsFeatureGloballyEnabledRequestPayload
	if err := jsondec.Decode(&payload); err != nil {
		return err
	}

	*featureName = payload.Feature
	return nil
}
