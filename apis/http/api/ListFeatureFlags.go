package api

import (
	"bytes"
	"encoding/json"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"log"
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

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(flags); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(`Content-Type`, `application/json`)
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Println(err)
	}
}
