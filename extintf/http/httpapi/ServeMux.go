package httpapi

import (
	"bytes"
	"encoding/json"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"log"
	"net/http"
)

func NewServeMux(uc *usecases.UseCases) *ServeMux {
	mux := &ServeMux{
		ServeMux: http.NewServeMux(),
		UseCases: uc,
	}

	mux.Handle(`/list-feature-flags.json`, http.HandlerFunc(mux.ListFeatureFlags))
	mux.Handle(`/is-feature-enabled-for.json`, http.HandlerFunc(mux.IsFeatureEnabledFor))
	mux.Handle(`/is-feature-globally-enabled.json`, http.HandlerFunc(mux.IsFeatureGloballyEnabled))
	mux.Handle(`/set-pilot-enrollment-for-feature.json`, http.HandlerFunc(mux.SetPilotEnrollmentForFeature))

	return mux
}

type ServeMux struct {
	*http.ServeMux
	*usecases.UseCases
}

func serveJSON(w http.ResponseWriter, status int, data interface{}) {
	buf := bytes.NewBuffer([]byte{})

	if err := json.NewEncoder(buf).Encode(data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(status)

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Println(err)
	}
}
