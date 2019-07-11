package httpapi

import (
	context2 "context"
	"net/http"

	"github.com/adamluzsi/toggler/usecases"
)

func (sm *ServeMux) ListFeatureFlags(w http.ResponseWriter, r *http.Request) {

	pu := r.Context().Value(`ProtectedUsecases`).(*usecases.ProtectedUsecases)

	ffs, err := pu.ListFeatureFlags(context2.TODO())

	if handleError(w, err, http.StatusInternalServerError) {
		return
	}

	serveJSON(w, &ffs)

}
