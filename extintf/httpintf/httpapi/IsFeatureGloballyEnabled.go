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
	jsondec := json.NewDecoder(r.Body)

	var payload IsFeatureEnabledForReqBody

	if handleError(w, jsondec.Decode(&payload), http.StatusBadRequest) {
		return
	}

	enrollment, err := sm.UseCases.IsFeatureGloballyEnabled(payload.Feature)

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
