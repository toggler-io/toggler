package httpapi

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"

	"github.com/toggler-io/toggler/usecases"
)

func NewServeMux(uc *usecases.UseCases) *ServeMux {
	mux := &ServeMux{
		UseCases: uc,
		ServeMux: http.NewServeMux(),
		Upgrader: &websocket.Upgrader{},
	}

	featureAPI := buildReleasesAPI(mux)
	mux.Handle(`/client/config.json`, http.HandlerFunc(mux.ClientConfigJSON))
	mux.Handle(`/release/`, http.StripPrefix(`/release`, featureAPI))
	mux.Handle(`/ws`, authMiddleware(uc, http.HandlerFunc(mux.WebsocketHandler)))

	mux.HandleFunc(`/healthcheck`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	return mux
}

func buildReleasesAPI(handlers *ServeMux) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle(`/is-feature-enabled.json`, http.HandlerFunc(handlers.IsFeatureEnabledFor))
	mux.Handle(`/is-feature-globally-enabled.json`, http.HandlerFunc(handlers.IsFeatureGloballyEnabled))
	mux.Handle(`/flag/`, http.StripPrefix(`/flag`, buildFlagAPI(handlers)))
	return mux
}

func buildFlagAPI(handlers *ServeMux) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(`/create.form`, http.HandlerFunc(handlers.CreateRolloutFeatureFlagFORM))
	mux.Handle(`/create.json`, http.HandlerFunc(handlers.CreateRolloutFeatureFlagJSON))
	mux.Handle(`/update.form`, http.HandlerFunc(handlers.UpdateFeatureFlagFORM))
	mux.Handle(`/update.json`, http.HandlerFunc(handlers.UpdateFeatureFlagJSON))
	mux.Handle(`/list.json`, http.HandlerFunc(handlers.ListFeatureFlags))
	mux.Handle(`/set-enrollment-manually.json`, http.HandlerFunc(handlers.SetPilotEnrollmentForFeature))
	return authMiddleware(handlers.UseCases, mux)
}

type ServeMux struct {
	*http.ServeMux
	*usecases.UseCases
	*websocket.Upgrader
}

func serveJSON(w http.ResponseWriter, data interface{}) {
	buf := bytes.NewBuffer([]byte{})

	if err := json.NewEncoder(buf).Encode(data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(200)

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Println(err)
	}
}
