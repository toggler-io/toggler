package controllers

import (
	"github.com/adamluzsi/toggler/extintf/httpintf/httputils"
	"github.com/adamluzsi/toggler/services/rollouts"
	"net/http"
	"net/url"
	"strings"
)

func (ctrl *Controller) PilotPage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case `/pilot/edit`:
		ctrl.pilotEditPage(w, r)

	case `/pilot/flag/set-rollout`:
		ctrl.pilotFlagSetRollout(w, r)

	default:
		http.NotFound(w, r)
	}
}

type pilotEditPageContent struct {
	Pilot        rollouts.Pilot
	FeatureFlags []struct {
		FeatureFlagID string
		PilotState    string
	}
}

func (ctrl *Controller) pilotEditPage(w http.ResponseWriter, r *http.Request) {
	if ctrl.handleError(w, r, r.ParseForm()) {
		return
	}

	//pilotExtID := r.URL.Query().Get(`ext-id`)

	//x := ctrl.GetProtectedUsecases(r).RolloutManager.FindPilotEntriesByExtID(r.Context(), pilotExtID)

	ctrl.Render(w, `/flag/show.html`, pilotEditPageContent{})
}


func (ctrl *Controller) pilotFlagSetRollout(w http.ResponseWriter, r *http.Request) {

	var pilot rollouts.Pilot
	pilot.ID = r.FormValue(`pilot.id`)
	pilot.FeatureFlagID = r.FormValue(`pilot.flagID`)
	pilot.ExternalID = r.FormValue(`pilot.extID`)

	switch strings.ToLower(r.FormValue(`pilot.enrollment`)) {
	case `whitelisted`:
		err := ctrl.GetProtectedUsecases(r).SetPilotEnrollmentForFeature(r.Context(), pilot.FeatureFlagID, pilot.ID, true)
		if httputils.HandleError(w, err, http.StatusInternalServerError) {
			return
		}

	case `blacklisted`:
		err := ctrl.GetProtectedUsecases(r).SetPilotEnrollmentForFeature(r.Context(), pilot.FeatureFlagID, pilot.ID, false)
		if httputils.HandleError(w, err, http.StatusInternalServerError) {
			return
		}

	case `undefined`:
		err := ctrl.GetProtectedUsecases(r).RolloutManager.Storage.DeleteByID(r.Context(), pilot, pilot.ID)
		if httputils.HandleError(w, err, http.StatusInternalServerError) {
			return
		}

	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	u, _ := url.Parse(`/pilot/edit`)
	q := u.Query()
	q.Set(`ext-id`, pilot.ExternalID)
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)

}