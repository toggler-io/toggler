package httpapi

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/adamluzsi/gorest"
	"github.com/gorilla/websocket"

	"github.com/toggler-io/toggler/extintf/httpintf/httputils"
	"github.com/toggler-io/toggler/usecases"
)

func NewHandler(uc *usecases.UseCases) *Handler {
	mux := &Handler{
		UseCases: uc,
		ServeMux: http.NewServeMux(),
		Upgrader: &websocket.Upgrader{},
	}

	gorest.Mount(mux.ServeMux, `/v`, NewViewsHandler(uc))

	gorest.Mount(mux.ServeMux, `/release-flags`,
		httputils.AuthMiddleware(gorest.NewHandler(ReleaseFlagController{UseCases: uc}), uc))

	featureAPI := buildReleasesAPI(mux)
	mux.Handle(`/release/`, http.StripPrefix(`/release`, featureAPI))

	mux.HandleFunc(`/healthcheck`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	return mux
}

func buildReleasesAPI(handlers *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle(`/is-feature-globally-enabled.json`, http.HandlerFunc(handlers.IsFeatureGloballyEnabled))
	mux.Handle(`/flag/`, http.StripPrefix(`/flag`, buildFlagAPI(handlers)))
	return mux
}

func buildFlagAPI(handlers *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(`/update.form`, http.HandlerFunc(handlers.UpdateFeatureFlagFORM))
	mux.Handle(`/update.json`, http.HandlerFunc(handlers.UpdateFeatureFlagJSON))
	mux.Handle(`/list.json`, http.HandlerFunc(handlers.ListFeatureFlags))
	mux.Handle(`/set-enrollment-manually.json`, http.HandlerFunc(handlers.SetPilotEnrollmentForFeature))
	return httputils.AuthMiddleware(mux, handlers.UseCases)
}

type Handler struct {
	*http.ServeMux
	*usecases.UseCases
	*websocket.Upgrader
}

func serveJSON(w http.ResponseWriter, data interface{}) {
	buf := bytes.NewBuffer([]byte{})

	if err := json.NewEncoder(buf).Encode(data); err != nil {
		log.Println(`ERROR`, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(200)

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Println(err)
	}
}
