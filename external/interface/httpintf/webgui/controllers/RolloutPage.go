package controllers

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/pkg/errors"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
)

func (ctrl *Controller) RolloutPage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case `/rollout`:
		ctrl.rolloutLandingPage(w, r)

	case `/rollout/index`:
		ctrl.rolloutIndexPage(w, r)

	case `/rollout/edit`:
		ctrl.rolloutEditPage(w, r)

	case `/rollout/update`:
		ctrl.rolloutUpdateAction(w, r)

	default:
		http.NotFound(w, r)
	}
}

func (ctrl *Controller) rolloutLandingPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		type Content struct {
			Environments []release.Environment
		}
		var content Content
		if ctrl.handleError(w, r, iterators.Collect(ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).FindAll(r.Context()), &content.Environments)) {
			return
		}
		ctrl.Render(w, `/rollout/landing.html`, content)

	case http.MethodPost:
		extID := r.FormValue(`rollout.flag_id`)
		envID := r.FormValue(`rollout.env_id`)

		u, _ := url.Parse(`/rollout/index`)
		q := u.Query()
		q.Set(`ext-id`, extID)
		q.Set(`env-id`, envID)
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)

	default:
		http.NotFound(w, r)

	}
}

func (ctrl *Controller) rolloutIndexPage(w http.ResponseWriter, r *http.Request) {
	if ctrl.handleError(w, r, r.ParseForm()) {
		return
	}

	envID := r.FormValue(`env-id`)
	if envID == `` {
		envID = r.URL.Query().Get(`env-id`)
	}

	var env release.Environment
	found, err := ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).FindByID(r.Context(), &env, envID)
	if httputils.HandleError(w, err, http.StatusNotFound) {
		log.Println(`ERROR`, err.Error())
		return
	}
	if !found {
		log.Println(`WARNING`, `environment id not found: `, envID)
		http.Redirect(w, r, `/`, http.StatusFound)
		return
	}

	ffs, err := ctrl.UseCases.RolloutManager.ListFeatureFlags(r.Context())

	if httputils.HandleError(w, err, http.StatusInternalServerError) {
		return
	}

	type ContentFeatureFlag struct {
		ReleaseFlagName   string
		ReleaseFlagID     string
		ReleasePercentage int
	}

	type Content struct {
		DeployEnvironmentID string
		FeatureFlags        []ContentFeatureFlag
	}
	var content Content
	content.DeployEnvironmentID = envID

	for _, ff := range ffs {
		var editFF ContentFeatureFlag
		editFF.ReleaseFlagID = ff.ID
		editFF.ReleaseFlagName = ff.Name

		var rollout release.Rollout
		var byPercentage = release.NewRolloutDecisionByPercentage()
		if found, err := ctrl.Storage.ReleaseRollout(r.Context()).FindByFlagEnvironment(r.Context(), ff, env, &rollout); ctrl.handleError(w, r, err) {
			return
		} else if found {
			if bp, ok := rollout.Plan.(release.RolloutDecisionByPercentage); ok {
				byPercentage = bp
			} else {
				log.Println(`ERROR`, `webgui is unable to handle the management of a complex rollout plan`)
				http.Redirect(w, r, `/`, http.StatusFound)
				return
			}
		}
		editFF.ReleasePercentage = byPercentage.Percentage

		content.FeatureFlags = append(content.FeatureFlags, editFF)
	}

	ctrl.Render(w, `/rollout/index.html`, content)
}

func (ctrl *Controller) rolloutEditPage(w http.ResponseWriter, r *http.Request) {
	if ctrl.handleError(w, r, r.ParseForm()) {
		return
	}

	query := r.URL.Query()
	flagID := query.Get(`flag-id`)
	envID := query.Get(`env-id`)

	log.Println(`flagID:`, flagID, `envID:`, envID)

	var env release.Environment
	if found, err := ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).FindByID(r.Context(), &env, envID); ctrl.handleError(w, r, err) {
		return
	} else if !found {
		http.Redirect(w, r, `/rollout`, http.StatusFound)
		return
	}

	var redirectToIndexPage = func() {
		u, _ := url.Parse(`/rollout/index`)
		q := u.Query()
		q.Set(`env-id`, envID)
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)
	}

	var flag release.Flag
	if found, err := ctrl.UseCases.Storage.ReleaseFlag(r.Context()).FindByID(r.Context(), &flag, flagID); ctrl.handleError(w, r, err) {
		return
	} else if !found {
		redirectToIndexPage()
		return
	}

	var byPercentage = release.NewRolloutDecisionByPercentage()

	var rollout release.Rollout
	if found, err := ctrl.Storage.ReleaseRollout(r.Context()).FindByFlagEnvironment(r.Context(), flag, env, &rollout); ctrl.handleError(w, r, err) {
		return
	} else if found {
		if bp, ok := rollout.Plan.(release.RolloutDecisionByPercentage); ok {
			byPercentage = bp
		} else {
			log.Println(`ERROR`, `webgui is unable to handle the management of a complex rollout plan`)
			redirectToIndexPage()
			return
		}
	}

	type ContentFeatureFlag struct {
		PilotState string
	}

	type Content struct {
		ReleaseFlagName       string
		ReleaseFlagID         string
		DeployEnvironmentID   string
		DeployEnvironmentName string
		ByPercentage          release.RolloutDecisionByPercentage
	}
	content := Content{
		ReleaseFlagName:       flag.Name,
		ReleaseFlagID:         flag.ID,
		DeployEnvironmentID:   env.ID,
		DeployEnvironmentName: env.Name,
		ByPercentage:          byPercentage,
	}

	ctrl.Render(w, `/rollout/edit.html`, content)
}

func (ctrl *Controller) rolloutUpdateAction(w http.ResponseWriter, r *http.Request) {
	var rollout release.Rollout
	rollout.FlagID = r.FormValue(`flag_id`)
	rollout.EnvironmentID = r.FormValue(`env_id`)

	if storedRollout, found, err := ctrl.lookupRollout(r.Context(), rollout.FlagID, rollout.EnvironmentID); ctrl.handleError(w, r, err) {
		return
	} else if found {
		rollout = storedRollout
	}

	percentage, err := strconv.Atoi(r.FormValue(`percentage`))
	if ctrl.handleError(w, r, err) {
		return
	}

	seed, err := strconv.ParseInt(r.FormValue(`seed`), 10, 64)
	if ctrl.handleError(w, r, err) {
		return
	}

	byPercentage := release.NewRolloutDecisionByPercentage()
	byPercentage.Percentage = percentage
	byPercentage.Seed = seed
	rollout.Plan = byPercentage

	if rollout.ID == `` {
		if ctrl.handleError(w, r, ctrl.UseCases.Storage.ReleaseRollout(r.Context()).Create(r.Context(), &rollout)) {
			return
		}
	} else {
		if ctrl.handleError(w, r, ctrl.UseCases.Storage.ReleaseRollout(r.Context()).Update(r.Context(), &rollout)) {
			return
		}
	}

	u, _ := url.Parse(`/rollout/index`)
	q := u.Query()
	q.Set(`env-id`, rollout.EnvironmentID)
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)

}

func (ctrl *Controller) lookupRollout(ctx context.Context, flagID, envID string) (release.Rollout, bool, error) {
	s := ctrl.UseCases.Storage

	var flag release.Flag
	if found, err := s.ReleaseFlag(ctx).FindByID(ctx, &flag, flagID); err != nil {
		return release.Rollout{}, false, err
	} else if !found {
		return release.Rollout{}, false, errors.New(`flag not found`)
	}

	var env release.Environment
	if found, err := s.ReleaseEnvironment(ctx).FindByID(ctx, &env, envID); err != nil {
		return release.Rollout{}, false, err
	} else if !found {
		return release.Rollout{}, false, errors.New(`env not found`)
	}

	var rollout release.Rollout
	found, err := s.ReleaseRollout(ctx).FindByFlagEnvironment(ctx, flag, env, &rollout)
	if err != nil {
		return release.Rollout{}, false, err
	}

	return rollout, found, nil
}
