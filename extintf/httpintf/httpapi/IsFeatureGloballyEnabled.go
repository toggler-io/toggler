package httpapi

import (
	"encoding/json"
	"net/http"
)

type IsFeatureGloballyEnabledPayload struct {
	Feature string `json:"feature"`
}

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
	var payload IsFeatureGloballyEnabledPayload
	if err := jsondec.Decode(&payload); err != nil {
		return err
	}

	*featureName = payload.Feature
	return nil
}
