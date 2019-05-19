package api

import (
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
	"strconv"
)

func (sm *ServeMux) SetPilotEnrollmentForFeature(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get(`token`)
	featureFlagName := r.URL.Query().Get(`feature`)
	extPilotID := r.URL.Query().Get(`user-id`)
	enrollmentStr := r.URL.Query().Get(`enrollment`)

	enrollment, err := strconv.ParseBool(enrollmentStr)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = sm.UseCases.SetPilotEnrollmentForFeature(token, featureFlagName, extPilotID, enrollment)

	if err == usecases.ErrInvalidToken {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}
