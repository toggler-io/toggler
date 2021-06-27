package controllers

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/pkg/errors"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
)

func (ctrl *Controller) PilotPage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case `/pilot/find`:
		ctrl.pilotFindPage(w, r)

	case `/pilot/edit`:
		ctrl.pilotEditPage(w, r)

	case `/pilot/flag/set-rollout`:
		ctrl.pilotFlagSetRollout(w, r)

	default:
		http.NotFound(w, r)
	}
}

func (ctrl *Controller) pilotFindPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		type Content struct {
			Environments []release.Environment
		}
		var content Content
		iterators.Collect(ctrl.UseCases.Storage.ReleaseEnvironment(r.Context()).FindAll(r.Context()), &content.Environments)
		ctrl.Render(w, `/pilot/find.html`, content)

	case http.MethodPost:
		extID := r.FormValue(`pilot.ext_id`)
		envID := r.FormValue(`pilot.env_id`)

		u, _ := url.Parse(`/pilot/edit`)
		q := u.Query()
		q.Set(`ext-id`, extID)
		q.Set(`env-id`, envID)
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)

	default:
		http.NotFound(w, r)

	}
}

func (ctrl *Controller) pilotEditPage(w http.ResponseWriter, r *http.Request) {
	if ctrl.handleError(w, r, r.ParseForm()) {
		return
	}

	query := r.URL.Query()
	pilotExtID := query.Get(`ext-id`)
	envID := query.Get(`env-id`)

	pilots := ctrl.UseCases.Storage.ReleasePilot(r.Context()).FindByPublicID(r.Context(), pilotExtID)
	defer pilots.Close()
	pilots = iterators.Filter(pilots, func(p release.Pilot) bool { return p.EnvironmentID == envID })

	pilotsIndex := make(map[string]release.Pilot) // FlagID => ManualPilot

	for pilots.Next() {
		var p release.Pilot

		if httputils.HandleError(w, pilots.Decode(&p), http.StatusInternalServerError) {
			return
		}

		pilotsIndex[p.FlagID] = p
	}

	ffs, err := ctrl.UseCases.RolloutManager.ListFeatureFlags(r.Context())

	if httputils.HandleError(w, err, http.StatusInternalServerError) {
		return
	}

	type ContentFeatureFlag struct {
		ReleaseFlagName     string
		ReleaseFlagID       string
		DeployEnvironmentID string
		PilotState          string
	}

	type Content struct {
		DeployEnvID     string
		PilotExternalID string
		FeatureFlags    []ContentFeatureFlag
	}
	var content Content
	content.PilotExternalID = pilotExtID
	content.DeployEnvID = envID

	for _, ff := range ffs {
		var editFF ContentFeatureFlag
		editFF.ReleaseFlagID = ff.ID
		editFF.ReleaseFlagName = ff.Name

		p, ok := pilotsIndex[ff.ID]

		if !ok {
			editFF.PilotState = `undefined`
		} else if p.IsParticipating {
			editFF.PilotState = `whitelisted`
		} else {
			editFF.PilotState = `blacklisted`
		}

		content.FeatureFlags = append(content.FeatureFlags, editFF)
	}

	ctrl.Render(w, `/pilot/edit.html`, content)
}

func (ctrl *Controller) pilotFlagSetRollout(w http.ResponseWriter, r *http.Request) {

	var pilot release.Pilot
	pilot.FlagID = r.FormValue(`pilot.flag_id`)
	pilot.EnvironmentID = r.FormValue(`pilot.env_id`)
	pilot.PublicID = r.FormValue(`pilot.ext_id`)
	newEnrollmentStatus := r.FormValue(`pilot.is_participating`)

	log.Println(`flag:`, pilot.FlagID,
		`env:`, pilot.EnvironmentID,
		`ext:`, pilot.PublicID,
		`is_participating`, newEnrollmentStatus)

	err := ctrl.setPilotManualEnrollmentForFlag(r.Context(), newEnrollmentStatus, pilot.FlagID, pilot.EnvironmentID, pilot.PublicID)

	if httputils.HandleError(w, err, http.StatusInternalServerError) {
		log.Println(err.Error())
	}

	u, _ := url.Parse(`/pilot/edit`)
	q := u.Query()
	q.Set(`ext-id`, pilot.PublicID)
	q.Set(`env-id`, pilot.EnvironmentID)
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)

}

func (ctrl *Controller) setPilotManualEnrollmentForFlag(ctx context.Context, newEnrollmentStatus, flagID, envID, extID string) error {
	var rm = ctrl.UseCases.RolloutManager
	switch strings.ToLower(newEnrollmentStatus) {
	case `whitelisted`:
		return rm.SetPilotEnrollmentForFeature(ctx, flagID, envID, extID, true)

	case `blacklisted`:
		return rm.SetPilotEnrollmentForFeature(ctx, flagID, envID, extID, false)

	case `undefined`:
		return rm.UnsetPilotEnrollmentForFeature(ctx, flagID, envID, extID)

	default:
		return errors.New(http.StatusText(http.StatusBadRequest))

	}
}

func ParseReleasePilotForm(r *http.Request) (*release.Pilot, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	var pilot release.Pilot
	pilot.ID = r.FormValue(`pilot.id`)
	pilot.FlagID = r.FormValue(`pilot.flag_id`)
	pilot.EnvironmentID = r.FormValue(`pilot.env_id`)
	pilot.PublicID = r.FormValue(`pilot.ext_id`)

	switch strings.ToLower(r.FormValue(`pilot.is_participating`)) {
	case `true`, `on`:
		pilot.IsParticipating = true
	case `false`, `off`, ``:
		pilot.IsParticipating = false
	default:
		return nil, errors.New(`unrecognised value for "pilot.is_participating" value`)
	}

	return &pilot, nil
}
