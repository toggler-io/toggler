package httpapi

import (
	"net/http"
)

func (sm *ServeMux) ListFeatureFlags(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get(`token`)

	pu, err := sm.UseCases.ProtectedUsecases(token)

	if errorHandler(w, err, http.StatusInternalServerError) {
		return
	}

	ffs, err := pu.ListFeatureFlags()

	if errorHandler(w, err, http.StatusInternalServerError) {
		return
	}

	serveJSON(w, 200, &ffs)
}
