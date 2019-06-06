package controllers

import (
	"net/http"
)

func (ctrl *Controller) IndexPage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case `/`, `/index`:
		flags, err := ctrl.GetProtectedUsecases(r).ListFeatureFlags()

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ctrl.Render(w, `/index.html`, flags)

	default:
		http.NotFound(w, r)
	}
}
