package api

import "net/http"

func (sm *ServeMux) IsFeatureEnabledFor(w http.ResponseWriter, r *http.Request) {

	values := r.URL.Query()
	featureFlagName := values.Get(`feature`)
	pilotID := values.Get(`user-id`)

	enrollment, err := sm.UseCases.IsFeatureEnabledFor(featureFlagName, pilotID)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var statusCode int

	if enrollment {
		statusCode = 200
	} else {
		statusCode = 403
	}

	w.WriteHeader(statusCode)

}
