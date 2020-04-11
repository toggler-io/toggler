package controllers

import (
	"log"
	"net/http"

	"github.com/toggler-io/toggler/external/interface/httpintf/webgui/cookies"
)

func (ctrl *Controller) LoginPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ctrl.Render(w, `/login.html`, nil)

	case http.MethodPost:

		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		token := r.FormValue(`token`)

		valid, err := ctrl.UseCases.Doorkeeper.VerifyTextToken(r.Context(), token)
		if err != nil {
			log.Println(`ERROR`, err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Redirect(w, r, `/login`, http.StatusFound)
			return
		}

		cookies.SetAuthToken(w, []byte(token))
		http.Redirect(w, r, `/`, http.StatusFound)

	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}
