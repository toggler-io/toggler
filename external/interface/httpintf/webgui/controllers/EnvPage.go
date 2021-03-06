package controllers

import (
	"fmt"
	"github.com/toggler-io/toggler/domains/release"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/adamluzsi/frameless/iterators"
)

func (ctrl *Controller) EnvPage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case `/env`:
		ctrl.envAction(w, r)
	case `/env/index`:
		ctrl.envListAction(w, r)
	case `/env/create`:
		ctrl.envCreateNewAction(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (ctrl *Controller) envListAction(w http.ResponseWriter, r *http.Request) {
	envsIter := ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).FindAll(r.Context())

	var envs []release.Environment
	if err := iterators.Collect(envsIter, &envs); err != nil {
		log.Println(`ERROR`, err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctrl.Render(w, `/env/index.html`, envs)
}

func (ctrl *Controller) envAction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if ctrl.handleError(w, r, r.ParseForm()) {
			return
		}

		id := r.Form.Get(`id`)

		var env release.Environment
		found, err := ctrl.UseCases.Storage.ReleaseRollout(r.Context()).FindByID(r.Context(), &env, id)

		if ctrl.handleError(w, r, err) {
			return
		}

		if !found {
			http.Redirect(w, r, `/`, http.StatusFound)
			return
		}

		type content struct {
			Env release.Environment
		}
		ctrl.Render(w, `/env/show.html`, content{Env: env})

	case http.MethodPost:
		switch strings.ToUpper(r.FormValue(`_method`)) {
		case http.MethodPut:
			env, err := ParseEnvForm(r)

			if ctrl.handleError(w, r, err) {
				return
			}

			if ctrl.handleError(w, r, ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).Update(r.Context(), &env)) {
				return
			}

			u, err := url.Parse(`/env`)

			if ctrl.handleError(w, r, err) {
				return
			}

			q := u.Query()
			q.Add(`id`, env.ID)
			u.RawQuery = q.Encode()
			http.Redirect(w, r, u.String(), http.StatusFound)
			return

		case http.MethodPost:
			env, err := ParseEnvForm(r)

			if ctrl.handleError(w, r, err) {
				return
			}

			if ctrl.handleError(w, r, ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).Create(r.Context(), &env)) {
				return
			}

			http.Redirect(w, r, `/env/index`, http.StatusFound)
			return

		case http.MethodDelete:
			if ctrl.handleError(w, r, r.ParseForm()) {
				return
			}

			envID := r.Form.Get(`env.id`)

			if envID == `` && ctrl.handleError(w, r, fmt.Errorf(`environment id is missing`)) {
				return
			}

			if ctrl.handleError(w, r, ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).DeleteByID(r.Context(), envID)) {
				return
			}

			http.Redirect(w, r, `/env/index`, http.StatusFound)
			return

		default:
			http.NotFound(w, r)
			return

		}

	default:
		http.NotFound(w, r)

	}
}

func (ctrl *Controller) envSetPilotAction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		p, err := ParseReleasePilotForm(r)

		if ctrl.handleError(w, r, err) {
			return
		}

		if ctrl.handleError(w, r, ctrl.UseCases.RolloutManager.SetPilotEnrollmentForFeature(r.Context(), p.FlagID, "", p.PublicID, p.IsParticipating)) {
			return
		}

		u, _ := url.Parse(`/env`)
		q := u.Query()
		q.Set(`id`, p.FlagID)
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)

	default:
		http.NotFound(w, r)

	}
}

func (ctrl *Controller) envCreateNewAction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ctrl.Render(w, `/env/create.html`, nil)

	case http.MethodPost:
		env, err := ParseEnvForm(r)

		if err != nil {
			log.Println(err)
			http.Redirect(w, r, `/`, http.StatusFound)
			return
		}

		if env.ID != `` {
			log.Println(`unexpected env id received`)
			http.Redirect(w, r, `/`, http.StatusFound)
			return
		}

		if env.Name == `` {
			log.Println(`missing env name`)
			http.Redirect(w, r, `/env/create`, http.StatusFound)
			return
		}

		err = ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).Create(r.Context(), &env)

		if err != nil {
			log.Println(err)
		}

		http.Redirect(w, r, `/`, http.StatusFound)

	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}

func ParseEnvForm(r *http.Request) (release.Environment, error) {
	if err := r.ParseForm(); err != nil {
		return release.Environment{}, err
	}
	var env release.Environment
	env.ID = r.Form.Get(`env.id`)
	env.Name = r.Form.Get(`env.name`)
	return env, nil
}
