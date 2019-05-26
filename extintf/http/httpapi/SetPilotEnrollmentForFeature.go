package httpapi

import (
	"net/http"
	"strconv"
)

func (sm *ServeMux) SetPilotEnrollmentForFeature(w http.ResponseWriter, r *http.Request) {

	token := r.URL.Query().Get(`token`)
	pu, err := sm.UseCases.ProtectedUsecases(token)

	if errorHandler(w, err, http.StatusInternalServerError) {
		return
	}

	featureFlagName := r.URL.Query().Get(`feature`)
	extPilotID := r.URL.Query().Get(`id`)
	enrollmentStr := r.URL.Query().Get(`enrollment`)
	enrollment, err := strconv.ParseBool(enrollmentStr)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = pu.SetPilotEnrollmentForFeature(featureFlagName, extPilotID, enrollment)

	if errorHandler(w, err, http.StatusInternalServerError) {
		return
	}

	serveJSON(w, 200, map[string]interface{}{})
}
