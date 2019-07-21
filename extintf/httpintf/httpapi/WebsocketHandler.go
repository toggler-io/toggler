package httpapi

import (
	"fmt"
	"github.com/pkg/errors"
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
	// Operation describe the chosen operation that needs to be executed.
	// required: true
	// enum: IsFeatureEnabled,IsFeatureGloballyEnabled
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

	handle := func(err error, code int) bool {
		if err == nil {
			return false
		}
		var errResp ErrorResponseBody
		errResp.Error.Code = code
		errResp.Error.Message = err.Error()
		return c.WriteJSON(errResp) != nil
	}

subscription:
	for {

		var req WebsocketRequestPayload

		if err := c.ReadJSON(&req); err != nil {
			break // err from Read is permanent
		}

		switch req.Operation {
		case `IsFeatureEnabled`:
			data := req.Data.(map[string]interface{})
			enr, err := sm.UseCases.IsFeatureEnabledFor(data[`feature`].(string), data[`id`].(string))
			if handle(err, http.StatusInternalServerError) {
				continue subscription
			}

			var resp IsFeatureEnabledResponseBody
			resp.Enrollment = enr

			if handle(c.WriteJSON(&resp), http.StatusInternalServerError) {
				break subscription
			}

		case `IsFeatureGloballyEnabled`:
			data := req.Data.(map[string]interface{})
			enr, err := sm.UseCases.IsFeatureGloballyEnabled(data[`feature`].(string))
			if handle(err, http.StatusInternalServerError) {
				continue subscription
			}

			var resp IsFeatureGloballyEnabledResponseBody
			resp.Enrollment = enr

			fmt.Println(resp)
			if handle(c.WriteJSON(&resp), http.StatusInternalServerError) {
				break subscription
			}

		default:
			if handle(errors.New(http.StatusText(http.StatusNotFound)), http.StatusNotFound) {
				break subscription
			}
		}

	}
}
