package webgui

import (
	"net/http"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/interface/httpintf/webgui/assets"
	"github.com/toggler-io/toggler/external/interface/httpintf/webgui/controllers"
)

func NewServeMux(uc *toggler.UseCases) (*ServeMux, error) {
	ctrl, err := controllers.NewController(uc)
	if err != nil {
		return nil, err
	}

	mux := &ServeMux{ServeMux: http.NewServeMux(), UseCases: uc}
	mux.Handle(`/assets/`, http.StripPrefix(`/assets`, assetsFS()))
	mux.HandleFunc(`/`, ctrl.IndexPage)
	mux.HandleFunc(`/flag`, ctrl.FlagPage)
	mux.HandleFunc(`/flag/`, ctrl.FlagPage)
	mux.HandleFunc(`/env`, ctrl.EnvPage)
	mux.HandleFunc(`/env/`, ctrl.EnvPage)
	mux.HandleFunc(`/rollout`, ctrl.RolloutPage)
	mux.HandleFunc(`/rollout/`, ctrl.RolloutPage)
	mux.HandleFunc(`/docs/`, ctrl.DocsPage)
	mux.HandleFunc(`/docs/assets/`, ctrl.DocsAssets)
	mux.HandleFunc(`/pilot/`, ctrl.PilotPage)
	mux.HandleFunc(`/login`, ctrl.LoginPage)
	return mux, nil
}

type ServeMux struct {
	*http.ServeMux
	*toggler.UseCases
}

func assetsFS() http.Handler {
	return http.FileServer(assets.FS(false))
}
