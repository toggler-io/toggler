package httpws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/interface/httpintf/httpapi"
)

type Controller struct {
	UseCases *toggler.UseCases
	Upgrader *websocket.Upgrader
}

// WebsocketRequestPayload is the payload that is expected to be received in the websocket connection.
//
// swagger:parameters Websocket
type WebsocketRequestParameter struct {
	// in: body
	Body WebsocketRequestPayload
}

type WebsocketRequestPayload struct {
	// Operation describe the chosen operation that needs to be executed.
	// required: true
	// enum: IsFeatureEnabled,GetReleaseFlagGlobalStates
	// example: IsFeatureEnabled
	Operation string `json:"operation"`
	// Data content correspond with the api payloads of the given operations.
	// example: {"feature":"my-feature","id":"pilot-id-name"}
	Data interface{} `json:"data"`
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
type WSLoadBalanceErrResp httpapi.ErrorResponse

/*

	swagger:route GET / ws release flag pilot global server-side websocket Websocket

	Socket API to check release flag Status

	This endpoint currently meant to used by servers and not by clients.
	The  reason behind is that it is much more easy to calculate with server quantity,
	than with client quantity, and therefore the load balancing is much more deterministic for the service.
	The websocket based communication allows for servers to do low latency quick requests,
	which is ideal to check flag status for individual requests that the server receives.
	Because the nature of the persistent connection, TCP connection overhead is minimal.
	The endpoint able to serve back whether the feature for a given pilot id is enabled or not.
	The endpoint also able to serve back global flag state checks as well.
	The flag enrollment interpretation use the same logic as it is described in the documentation.

		Consumes:
		- application/json

		Produces:
		- application/json

		Security:
		  api_key:
		  - "X-Auth-Token"

		Responses:
		  200: enrollmentResponse
		  500: errorResponse
		  503: wsLoadBalanceErrResponse

*/
func (ctrl *Controller) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: 503 Service Unavailable for rand based load balancing

	c, err := ctrl.Upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	defer c.Close()

	handle := func(err error, code int) bool {
		if err == nil {
			return false
		}
		var errResp httpapi.ErrorResponse
		errResp.Body.Error.Code = code
		errResp.Body.Error.Message = err.Error()
		return c.WriteJSON(errResp.Body) != nil
	}

subscription:
	for {

		var req WebsocketRequestPayload

		if err := c.ReadJSON(&req); err != nil {
			break // err from Read is permanent
		}

		switch req.Operation {
		case `IsFeatureEnabled`:
			//data := req.Data.(map[string]interface{})
			//
			//releaseFlagName := data[`feature`].(string)
			//states, err := ctrl.UseCases.GetAllReleaseFlagStatesOfThePilot(r.Context(), releaseFlagName, data[`id`].(string))
			//
			//if handle(err, http.StatusInternalServerError) {
			//	continue subscription
			//}
			//
			//var resp EnrollmentResponseBody
			//resp.Enrollment = states[releaseFlagName]
			//
			//if handle(c.WriteJSON(&resp), http.StatusInternalServerError) {
			//	break subscription
			//}

		default:
			if handle(errors.New(http.StatusText(http.StatusNotFound)), http.StatusNotFound) {
				break subscription
			}
		}

	}
}

type IsFeatureEnabledRequestPayload struct {
	// name is the name Flag name that is needed to be checked for enrollment
	//
	// required: true
	// example: rollout-feature-flag
	Feature string `json:"feature"`
	// PilotExtID is the public unique ID of the pilot who's enrollment needs to be checked.
	//
	// required: true
	// example: pilot-public-id
	PilotID string `json:"id"`
}
