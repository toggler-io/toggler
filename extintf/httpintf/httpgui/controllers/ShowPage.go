package httpgui

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"net/http"
)

func (mux *ServeMux) ShowPage(w http.ResponseWriter, r *http.Request) {
	mux.render(w, `/show.html`, rollouts.FeatureFlag{})
}
