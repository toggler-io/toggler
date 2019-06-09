package controllers

import (
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
	"time"
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

		_, err := ctrl.ProtectedUsecases(token)

		if err == usecases.ErrInvalidToken {
			http.Redirect(w, r, `/login`, http.StatusFound)
			return
		}

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		expiration := time.Now().Add(365 * 24 * time.Hour)
		cookie := http.Cookie{Name: `token`, Value: token, Expires: expiration}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, `/`, http.StatusFound)

	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}
