package httputils

import (
	"net/http"
	"strings"

	"github.com/toggler-io/toggler/services/rollouts"
	"github.com/pkg/errors"
)

func ParseFlagPilotFromForm(r *http.Request) (*rollouts.Pilot, error) {

	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	var pilot rollouts.Pilot
	pilot.ID = r.FormValue(`pilot.id`)
	pilot.FeatureFlagID = r.FormValue(`pilot.flagID`)
	pilot.ExternalID = r.FormValue(`pilot.extID`)

	switch strings.ToLower(r.FormValue(`pilot.enrolled`)) {
	case `true`, `on`:
		pilot.Enrolled = true
	case `false`, ``:
		pilot.Enrolled = false
	default:
		return nil, errors.New(`unrecognised value for "pilot.enrolled" value`)
	}

	return &pilot, nil

}
