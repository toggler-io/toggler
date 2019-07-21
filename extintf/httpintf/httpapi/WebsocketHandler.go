package httpapi

import (
	"net/http"
)

// WebsocketRequestPayload is the payload that is expected to be received in the websocket connection.
//
// swagger:parameters Websocket
type WebsocketRequestParameter struct {
	// in: body
	Body WebsocketRequestPayload
}

type WebsocketRequestPayload struct {
	// Feature is the Feature Flag name that is needed to be checked for enrollment
	//
	// required: true
	// example: rollout-feature-flag
	Feature string  `json:"feature"`
	// PilotID is the public unique ID of the pilot who's enrollment needs to be checked.
	//
	// in: body
	// example: pilot-public-id
	PilotID *string `json:"id,omitempty"`
}

type WebsocketResponseBody = EnrollmentResponseBody

// WSLoadBalanceErrResp will be received in case the receiver server cannot take more ws connections.
// This error must be handled by retrying the call until it succeed.
// This error meant to be a recoverable error.
// The main purpose for this is to gain control over how  much ws connection can be open on a single server instance,
// so scaling the service can be easily achieved.
// In case there is a load balancer that handle this transparently, this error may not be received.
//
// swagger:response wsLoadBalanceErrResponse
type WSLoadBalanceErrResp struct {
	// Error contains the details of the error
	// in: body
	Body ErrorResponseBody
}


/*

	swagger:route GET /api/v1/ws feature-flag pilot global websocket Websocket

	Socket API to check Rollout Feature Flag Status

	The
	Reply back whether the feature for a given pilot id is enabled or not.
	By Default, this will be determined whether the flag exist,
	the pseudo random dice roll enrolls the pilot,
	or if there any manually set enrollment status for the pilot.

		Consumes:
		- application/json

		Produces:
		- application/json

		Schemes: ws

		Responses:
		  200: enrollmentResponse
		  500: errorResponse
		  503: wsLoadBalanceErrResponse

*/
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
