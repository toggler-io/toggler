package api

import (
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
)

func (sm *ServeMux) ListFeatureFlags(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get(`token`)

	flags, err := sm.UseCases.ListFeatureFlags(token)

	if err == usecases.ErrInvalidToken {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	serveJSON(w, 200, &flags)
}
