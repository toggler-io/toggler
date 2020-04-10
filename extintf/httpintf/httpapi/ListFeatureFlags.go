package httpapi

import (
	 "context"
	"net/http"

	"github.com/toggler-io/toggler/usecases"
)

func (sm *Handler) ListFeatureFlags(w http.ResponseWriter, r *http.Request) {

	pu := r.Context().Value(`ProtectedUseCases`).(*usecases.ProtectedUseCases)

	ffs, err := pu.ListFeatureFlags(context.TODO())

	if handleError(w, err, http.StatusInternalServerError) {
		return
	}

	serveJSON(w, &ffs)

}
