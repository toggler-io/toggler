package httpapi

import (
	"net/http"
)

type WebsocketRequestPayload struct {
	Feature string  `json:"feature"`
	PilotID *string `json:"id,omitempty"`
}

type WebsocketResponseBody = EnrollmentResponseBody

func (sm *ServeMux) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: 503 Service Unavailable for rand based load balancing

	c, err := sm.Upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	defer c.Close()

subscription:
	for {
		var req WebsocketRequestPayload

		if err := c.ReadJSON(&req); err != nil {
			break // err from Read is permanent
		}

		var res WebsocketResponseBody

		var enr bool
		if req.PilotID == nil {
			enr, err = sm.UseCases.IsFeatureGloballyEnabled(req.Feature)
		} else {
			enr, err = sm.UseCases.IsFeatureEnabledFor(req.Feature, *req.PilotID)
		}

		if err != nil {
			var errResp ErrorResponseBody
			errResp.Error.Code = http.StatusInternalServerError
			errResp.Error.Message = err.Error()
			if werr := c.WriteJSON(errResp); werr != nil {
				break subscription
			}
			continue subscription
		}

		res.Enrollment = enr

		if err := c.WriteJSON(res); err != nil {
			break
		}
	}
}
