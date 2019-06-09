package httputils

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"net/http"
	"strconv"
)

func ParseFlagPilotFromForm(r *http.Request) (*rollouts.Pilot, error) {

	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	var pilot rollouts.Pilot

	pilot.ID = r.FormValue(`pilot.id`)
	pilot.FeatureFlagID = r.FormValue(`pilot.flagID`)
	pilot.ExternalID = r.FormValue(`pilot.extID`)
	enrollment, err := strconv.ParseBool(r.FormValue(`pilot.enrolled`))

	if err != nil {
		return nil, err
	}

	pilot.Enrolled = enrollment

	return &pilot, nil

}
