package controllers

import (
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httputils"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/frameless/iterators"
	"log"
	"net/http"
)

type edigPageContent struct {
	Flag   rollouts.FeatureFlag
	Pilots []rollouts.Pilot
}

// super hacky implementation for the sake of POC
func (ctrl *Controller) EditPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if ctrl.handleError(w, r, r.ParseForm()) {
			return
		}

		id := r.Form.Get(`id`)

		var ff rollouts.FeatureFlag
		found, err := ctrl.Storage.FindByID(id, &ff)

		if ctrl.handleError(w, r, err) {
			return
		}

		if !found {
			http.Redirect(w, r, `/`, http.StatusFound)
			return
		}

		var pilots []rollouts.Pilot

		if ctrl.handleError(w, r, iterators.CollectAll(ctrl.Storage.FindPilotsByFeatureFlag(&ff), &pilots)) {
			return
		}

		ctrl.Render(w, `/edit.html`, edigPageContent{Flag: ff, Pilots: pilots})

	case http.MethodPost:
		ff, err := httputils.ParseFlagFromForm(r)

		if ctrl.handleError(w, r, err) {
			return
		}

		if ff.ID == `` {
			log.Println(`missing flag id`)
			http.Redirect(w, r, `/`, http.StatusFound)
			return
		}

		err = ctrl.GetProtectedUsecases(r).UpdateFeatureFlag(ff)

		if err != nil {
			log.Println(err)
		}

		http.Redirect(w, r, `/edit`, http.StatusFound)

	default:
		http.Redirect(w, r, `/`, http.StatusFound)

	}
}

func (ctrl *Controller) EditEnrollmentFormAction(w http.ResponseWriter, r *http.Request) {

}
