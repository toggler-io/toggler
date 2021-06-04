package httpapi

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/adamluzsi/gorest"
	"github.com/gorilla/websocket"

	"github.com/toggler-io/toggler/domains/toggler"
)

func NewHandler(uc *toggler.UseCases) *Handler {
	mux := &Handler{
		UseCases: uc,
		ServeMux: http.NewServeMux(),
		Upgrader: &websocket.Upgrader{},
	}

	gorest.Mount(mux.ServeMux, `/v`, NewViewsHandler(uc))
	gorest.Mount(mux.ServeMux, `/release-flags`, NewReleaseFlagHandler(uc))
	gorest.Mount(mux.ServeMux, `/deployment-environments`, NewDeploymentEnvironmentHandler(uc))
	gorest.Mount(mux.ServeMux, `/release-pilots`, NewReleasePilotHandler(uc))
	gorest.Mount(mux.ServeMux, `/release-rollouts`, NewReleaseRolloutHandler(uc))

	mux.HandleFunc(`/healthcheck`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	return mux
}

type Handler struct {
	*http.ServeMux
	*toggler.UseCases
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
