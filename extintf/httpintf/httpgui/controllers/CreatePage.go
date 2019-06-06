package controllers

import (
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httputils"
	"log"
	"net/http"
)

func (ctrl *Controller) CreatePage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ctrl.Render(w, `/create.html`, nil)

	case http.MethodPost:

		ff, err := httputils.ParseFlagFromForm(r)

		if err != nil {
			log.Println(err)
			http.Redirect(w, r, `/`, http.StatusFound)
			return
		}

		if ff.ID != `` {
			log.Println(`unexpected flag id received`)
			http.Redirect(w, r, `/`, http.StatusFound)
			return
		}

		err = ctrl.GetProtectedUsecases(r).CreateFeatureFlag(ff)

		if err != nil {
			log.Println(err)
		}

		http.Redirect(w, r, `/`, http.StatusFound)

	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}
