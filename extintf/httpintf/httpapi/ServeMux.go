package httpapi

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"

	"github.com/adamluzsi/toggler/usecases"
)

func NewServeMux(uc *usecases.UseCases) *ServeMux {
	mux := &ServeMux{
		UseCases: uc,
		ServeMux: http.NewServeMux(),
		Upgrader: &websocket.Upgrader{},
	}

	featureAPI := buildFeatureAPI(mux)
	mux.Handle(`/rollout/`, http.StripPrefix(`/rollout`, featureAPI))
	mux.HandleFunc(`/ws`, mux.WebsocketHandler)

	return mux
}

func buildFeatureAPI(handlers *ServeMux) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle(`/config.json`, http.HandlerFunc(handlers.RolloutConfigJSON))
	mux.Handle(`/is-enabled.json`, http.HandlerFunc(handlers.IsFeatureEnabledFor))
	mux.Handle(`/is-globally-enabled.json`, http.HandlerFunc(handlers.IsFeatureGloballyEnabled))
	mux.Handle(`/flag/`, http.StripPrefix(`/flag`, buildFlagAPI(handlers)))
	return mux
}

func buildFlagAPI(handlers *ServeMux) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(`/create.form`, http.HandlerFunc(handlers.CreateFeatureFlagFORM))
	mux.Handle(`/create.json`, http.HandlerFunc(handlers.CreateFeatureFlagJSON))
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
