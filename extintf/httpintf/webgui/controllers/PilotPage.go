package controllers

import (
	"context"
	"github.com/adamluzsi/toggler/extintf/httpintf/httputils"
	"github.com/adamluzsi/toggler/services/rollouts"
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
	FeatureFlagID string
	PilotState    string
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

	protectedUsecases := ctrl.GetProtectedUsecases(r)

	pilots := protectedUsecases.RolloutManager.FindPilotEntriesByExtID(r.Context(), pilotExtID)

	pilotsIndex := make(map[string]rollouts.Pilot) // FlagID => Pilot

	for pilots.Next() {
		var p rollouts.Pilot

		if httputils.HandleError(w, pilots.Decode(&p), http.StatusInternalServerError) {
			return
		}

		pilotsIndex[p.FeatureFlagID] = p
	}

	ffs, err := protectedUsecases.RolloutManager.ListFeatureFlags(r.Context())

	if httputils.HandleError(w, err, http.StatusInternalServerError) {
		return
	}

	var data pilotEditPageContent

	data.PilotExternalID = pilotExtID

	for _, ff := range ffs {
		var editFF pilotEditPageContentFeatureFlag
		editFF.FeatureFlagID = ff.ID

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

	var pilot rollouts.Pilot
	pilot.FeatureFlagID = r.FormValue(`pilot.flagID`)
	pilot.ExternalID = r.FormValue(`pilot.extID`)

	newEnrollmentStatus := strings.ToLower(r.FormValue(`pilot.enrollment`))
	log.Println(pilot.ExternalID, newEnrollmentStatus)

	err := ctrl.setPilotEnrollmentForFlag(
		r.Context(),
		ctrl.GetProtectedUsecases(r).RolloutManager,
		newEnrollmentStatus,
		pilot.FeatureFlagID,
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

func (ctrl *Controller) setPilotEnrollmentForFlag(ctx context.Context, rm *rollouts.RolloutManager, newEnrollmentStatus string, flagID, pilotExtID string) error {
	switch newEnrollmentStatus {
	case `whitelisted`:
		return rm.SetPilotEnrollmentForFeature(ctx, flagID, pilotExtID, true)

	case `blacklisted`:
		return rm.SetPilotEnrollmentForFeature(ctx, flagID, pilotExtID, false)

	case `undefined`:

		p, err := rm.FindFlagPilotByExternalPilotID(ctx, flagID, pilotExtID)

		if err != nil {
			return err
		}

		if p == nil {
			return nil
		}

		return rm.Storage.DeleteByID(ctx, rollouts.Pilot{}, p.ID)

	default:
		return errors.New(http.StatusText(http.StatusBadRequest))

	}
}
