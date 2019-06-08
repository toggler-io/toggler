package httpapi

import (
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
	"strconv"
)

func (sm *ServeMux) SetPilotEnrollmentForFeature(w http.ResponseWriter, r *http.Request) {

	pu := r.Context().Value(`ProtectedUsecases`).(*usecases.ProtectedUsecases)

	flagID := r.URL.Query().Get(`flagID`)
	pilotID := r.URL.Query().Get(`pilotID`)
	enrollment, err := strconv.ParseBool(r.URL.Query().Get(`enrolled`))

	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = pu.SetPilotEnrollmentForFeature(flagID, pilotID, enrollment)

	if handleError(w, err, http.StatusInternalServerError) {
		return
	}

	serveJSON(w, 200, map[string]interface{}{})
}
