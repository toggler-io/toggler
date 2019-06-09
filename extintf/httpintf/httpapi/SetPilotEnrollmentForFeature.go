package httpapi

import (
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httputils"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
)

func (sm *ServeMux) SetPilotEnrollmentForFeature(w http.ResponseWriter, r *http.Request) {

	pu := r.Context().Value(`ProtectedUsecases`).(*usecases.ProtectedUsecases)

	pilot, err := httputils.ParseFlagPilotFromForm(r)

	if err != nil || pilot.FeatureFlagID == `` || pilot.ExternalID == `` {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = pu.SetPilotEnrollmentForFeature(pilot.FeatureFlagID, pilot.ExternalID, pilot.Enrolled)

	if handleError(w, err, http.StatusInternalServerError) {
		return
	}

	serveJSON(w, 200, map[string]interface{}{})
}
