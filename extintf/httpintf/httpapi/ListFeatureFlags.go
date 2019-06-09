package httpapi

import (
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
)

func (sm *ServeMux) ListFeatureFlags(w http.ResponseWriter, r *http.Request) {

	pu := r.Context().Value(`ProtectedUsecases`).(*usecases.ProtectedUsecases)

	ffs, err := pu.ListFeatureFlags()

	if handleError(w, err, http.StatusInternalServerError) {
		return
	}

	serveJSON(w, 200, &ffs)
}
