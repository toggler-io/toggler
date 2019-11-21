package controllers

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"

	"github.com/toggler-io/toggler/extintf/httpintf/httputils"
	"github.com/toggler-io/toggler/services/release"
)

type editPageContent struct {
	Flag   release.Flag
	Pilots []release.Pilot
}

func (ctrl *Controller) FlagPage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case `/flag`:
		ctrl.flagAction(w, r)
	case `/flag/index`:
		ctrl.flagListAction(w, r)
	case `/flag/create`:
		ctrl.flagCreateNewAction(w, r)
	case `/flag/pilot`, `/flag/pilot/update`:
		ctrl.flagSetPilotAction(w, r)
	case `/flag/pilot/unset`:
		ctrl.flagUnsetPilotAction(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (ctrl *Controller) flagListAction(w http.ResponseWriter, r *http.Request) {
	flags, err := ctrl.GetProtectedUsecases(r).ListFeatureFlags(r.Context())

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctrl.Render(w, `/flag/index.html`, flags)
}

func (ctrl *Controller) flagAction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if ctrl.handleError(w, r, r.ParseForm()) {
			return
		}

		id := r.Form.Get(`id`)

		var ff release.Flag
		found, err := ctrl.Storage.FindByID(r.Context(), &ff, id)

		if ctrl.handleError(w, r, err) {
			return
		}

		if !found {
			http.Redirect(w, r, `/`, http.StatusFound)
			return
		}

		var pilots []release.Pilot

		if ctrl.handleError(w, r, iterators.Collect(ctrl.Storage.FindPilotsByFeatureFlag(r.Context(), &ff), &pilots)) {
			return
		}

		ctrl.Render(w, `/flag/show.html`, editPageContent{Flag: ff, Pilots: pilots})

	case http.MethodPost:
		switch strings.ToUpper(r.FormValue(`_method`)) {
		case http.MethodPut:
			ff, err := httputils.ParseFlagFromForm(r)

			if ctrl.handleError(w, r, err) {
				return
			}

			if ctrl.handleError(w, r, ctrl.GetProtectedUsecases(r).UpdateFeatureFlag(r.Context(), ff)) {
				return
			}

			u, err := url.Parse(`/flag`)

			if ctrl.handleError(w, r, err) {
				return
			}

			q := u.Query()
			q.Add(`id`, ff.ID)
			u.RawQuery = q.Encode()
			http.Redirect(w, r, u.String(), http.StatusFound)
			return

		case http.MethodPost:
			ff, err := httputils.ParseFlagFromForm(r)

			if ctrl.handleError(w, r, err) {
				return
			}

			if ctrl.handleError(w, r, ctrl.GetProtectedUsecases(r).CreateFeatureFlag(r.Context(), ff)) {
				return
			}

			u, err := url.Parse(`/flag`)

			if ctrl.handleError(w, r, err) {
				return
			}

			q := u.Query()
			q.Add(`id`, ff.ID)
			u.RawQuery = q.Encode()
			http.Redirect(w, r, u.String(), http.StatusFound)
			return

		case http.MethodDelete:
			if ctrl.handleError(w, r, r.ParseForm()) {
				return
			}

			flagID := r.Form.Get(`flag.id`)

			if flagID == `` && ctrl.handleError(w, r, frameless.ErrIDRequired) {
				return
			}

			if ctrl.handleError(w, r, ctrl.GetProtectedUsecases(r).DeleteFeatureFlag(r.Context(), flagID)) {
				return
			}

			http.Redirect(w, r, `/flag/index`, http.StatusFound)
			return

		default:
			http.NotFound(w, r)
			return

		}

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

		if ctrl.handleError(w, r, ctrl.GetProtectedUsecases(r).SetPilotEnrollmentForFeature(r.Context(), p.FlagID, p.ExternalID, p.Enrolled)) {
			return
		}

		u, _ := url.Parse(`/flag`)
		q := u.Query()
		q.Set(`id`, p.FlagID)
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)

	default:
		http.NotFound(w, r)

	}
}

func (ctrl *Controller) flagUnsetPilotAction(w http.ResponseWriter, r *http.Request) {
	featureFlagID := r.FormValue(`pilot.flagID`)
	pilotExternalID := r.FormValue(`pilot.extID`)

	err := ctrl.GetProtectedUsecases(r).UnsetPilotEnrollmentForFeature(r.Context(), featureFlagID, pilotExternalID)

	if ctrl.handleError(w, r, err) {
		return
	}

	u, _ := url.Parse(`/flag`)
	q := u.Query()
	q.Set(`id`, featureFlagID)
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
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

		err = ctrl.GetProtectedUsecases(r).CreateFeatureFlag(context.TODO(), ff)

		if err != nil {
			log.Println(err)
		}

		http.Redirect(w, r, `/`, http.StatusFound)

	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}
