package controllers

import (
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/toggler/extintf/httpintf/httputils"
	"github.com/adamluzsi/toggler/services/rollouts"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type edigPageContent struct {
	Flag   rollouts.FeatureFlag
	Pilots []rollouts.Pilot
}

func (ctrl *Controller) FlagPage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case `/flag`:
		ctrl.flagAction(w, r)
	case `/flag/create`:
		ctrl.flagCreateNewAction(w, r)
	case `/flag/pilot`, `/flag/pilot/update`:
		ctrl.flagSetPilotAction(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (ctrl *Controller) flagAction(w http.ResponseWriter, r *http.Request) {
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

		ctrl.Render(w, `/flag/show.html`, edigPageContent{Flag: ff, Pilots: pilots})

	case http.MethodPost:
		ff, err := httputils.ParseFlagFromForm(r)

		if ctrl.handleError(w, r, err) {
			return
		}

		if strings.ToLower(r.Form.Get(`_method`)) == `put` {
			if ctrl.handleError(w, r, ctrl.GetProtectedUsecases(r).UpdateFeatureFlag(ff)) {
				return
			}
		} else {
			if ff.ID != `` {
				log.Println(`unexpected flag id received`)
				http.Redirect(w, r, `/`, http.StatusFound)
				return
			}

			if ctrl.handleError(w, r, ctrl.GetProtectedUsecases(r).CreateFeatureFlag(ff)) {
				return
			}
		}
		http.Redirect(w, r, `/flag`, http.StatusFound)

	default:
		http.NotFound(w, r)

	}
}

func (ctrl *Controller) flagSetPilotAction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		p, err := httputils.ParseFlagPilotFromForm(r)

		if ctrl.handleError(w, r, err) {
			return
		}

		if ctrl.handleError(w, r, ctrl.GetProtectedUsecases(r).SetPilotEnrollmentForFeature(
			p.FeatureFlagID, p.ExternalID, p.Enrolled)) {
			return
		}

		u, _ := url.Parse(`/flag`)
		q := u.Query()
		q.Set(`id`, p.FeatureFlagID)
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)

	default:
		http.NotFound(w, r)

	}
}

func (ctrl *Controller) flagCreateNewAction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ctrl.Render(w, `/flag/create.html`, nil)

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

		if ff.Name == `` {
			log.Println(`missing flag name`)
			http.Redirect(w, r, `/flag/create`, http.StatusFound)
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
