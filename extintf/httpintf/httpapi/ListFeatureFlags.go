package httpapi

import (
	 "context"
	"net/http"

	"github.com/adamluzsi/toggler/usecases"
)

func (sm *ServeMux) ListFeatureFlags(w http.ResponseWriter, r *http.Request) {

	pu := r.Context().Value(`ProtectedUsecases`).(*usecases.ProtectedUsecases)

	ffs, err := pu.ListFeatureFlags(context.TODO())

	if handleError(w, err, http.StatusInternalServerError) {
		return
	}

	serveJSON(w, &ffs)

}
