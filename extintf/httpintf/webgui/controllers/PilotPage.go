package controllers

import (
	"context"
	"github.com/toggler-io/toggler/extintf/httpintf/httputils"
	"github.com/toggler-io/toggler/services/release"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"net/url"
	"strings"
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

type pilotEditPageContent struct {
	PilotExternalID string
	FeatureFlags    []pilotEditPageContentFeatureFlag
}

type pilotEditPageContentFeatureFlag struct {
	ReleaseFlagName string
	ReleaseFlagID   string
	PilotState      string
}

func (ctrl *Controller) pilotFindPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ctrl.Render(w, `/pilot/find.html`, nil)

	case http.MethodPost:
		pilotExtID := r.FormValue(`pilot.extID`)
		u, _ := url.Parse(`/pilot/edit`)
		q := u.Query()
		q.Set(`ext-id`, pilotExtID)
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

	pilotExtID := r.URL.Query().Get(`ext-id`)

	pilots := ctrl.UseCases.RolloutManager.FindPilotEntriesByExtID(r.Context(), pilotExtID)

	pilotsIndex := make(map[string]release.Pilot) // FlagID => Pilot

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

	var data pilotEditPageContent

	data.PilotExternalID = pilotExtID

	for _, ff := range ffs {
		var editFF pilotEditPageContentFeatureFlag
		editFF.ReleaseFlagID = ff.ID
		editFF.ReleaseFlagName = ff.Name

		p, ok := pilotsIndex[ff.ID]
		if !ok {
			editFF.PilotState = `undefined`
		} else if p.Enrolled {
			editFF.PilotState = `whitelisted`
		} else {
			editFF.PilotState = `blacklisted`
		}

		data.FeatureFlags = append(data.FeatureFlags, editFF)
	}

	ctrl.Render(w, `/pilot/edit.html`, data)
}

func (ctrl *Controller) pilotFlagSetRollout(w http.ResponseWriter, r *http.Request) {

	var pilot release.Pilot
	pilot.FlagID = r.FormValue(`pilot.flagID`)
	pilot.ExternalID = r.FormValue(`pilot.extID`)

	newEnrollmentStatus := strings.ToLower(r.FormValue(`pilot.enrollment`))
	log.Println(pilot.ExternalID, newEnrollmentStatus)

	err := ctrl.setPilotManualEnrollmentForFlag(
		r.Context(),
		newEnrollmentStatus,
		pilot.FlagID,
		pilot.ExternalID,
	)

	if httputils.HandleError(w, err, http.StatusInternalServerError) {
		log.Println(err.Error())
	}

	u, _ := url.Parse(`/pilot/edit`)
	q := u.Query()
	q.Set(`ext-id`, pilot.ExternalID)
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)

}

func (ctrl *Controller) setPilotManualEnrollmentForFlag(ctx context.Context, newEnrollmentStatus string, flagID, pilotExtID string) error {
	var rm = ctrl.UseCases.RolloutManager
	switch newEnrollmentStatus {
	case `whitelisted`:
		return rm.SetPilotEnrollmentForFeature(ctx, flagID, pilotExtID, true)

	case `blacklisted`:
		return rm.SetPilotEnrollmentForFeature(ctx, flagID, pilotExtID, false)

	case `undefined`:
		return rm.UnsetPilotEnrollmentForFeature(ctx, flagID, pilotExtID)

	default:
		return errors.New(http.StatusText(http.StatusBadRequest))

	}
}
