// Code generated by go-swagger; DO NOT EDIT.

package release_flag

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/toggler-io/toggler/lib/go/models"
)

// WebsocketReader is a Reader for the Websocket structure.
type WebsocketReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *WebsocketReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewWebsocketOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewWebsocketInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 503:
		result := NewWebsocketServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewWebsocketOK creates a WebsocketOK with default headers values
func NewWebsocketOK() *WebsocketOK {
	return &WebsocketOK{}
}

/*WebsocketOK handles this case with default header values.

EnrollmentResponse returns information about the requester's rollout feature enrollment status.
*/
type WebsocketOK struct {
	Payload *models.EnrollmentResponseBody
}

func (o *WebsocketOK) Error() string {
	return fmt.Sprintf("[GET /ws][%d] websocketOK  %+v", 200, o.Payload)
}

func (o *WebsocketOK) GetPayload() *models.EnrollmentResponseBody {
	return o.Payload
}

func (o *WebsocketOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.EnrollmentResponseBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewWebsocketInternalServerError creates a WebsocketInternalServerError with default headers values
func NewWebsocketInternalServerError() *WebsocketInternalServerError {
	return &WebsocketInternalServerError{}
}

/*WebsocketInternalServerError handles this case with default header values.

ErrorResponse will contains a response about request that had some kind of problem.
The details will be included in the body.
*/
type WebsocketInternalServerError struct {
	Payload *models.ErrorResponseBody
}

func (o *WebsocketInternalServerError) Error() string {
	return fmt.Sprintf("[GET /ws][%d] websocketInternalServerError  %+v", 500, o.Payload)
}

func (o *WebsocketInternalServerError) GetPayload() *models.ErrorResponseBody {
	return o.Payload
}

func (o *WebsocketInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponseBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewWebsocketServiceUnavailable creates a WebsocketServiceUnavailable with default headers values
func NewWebsocketServiceUnavailable() *WebsocketServiceUnavailable {
	return &WebsocketServiceUnavailable{}
}

/*WebsocketServiceUnavailable handles this case with default header values.

WSLoadBalanceErrResp will be received in case the receiver server cannot take more ws connections.
This error must be handled by retrying the call until it succeed.
This error meant to be a recoverable error.
The main purpose for this is to gain control over how  much ws connection can be open on a single server instance,
so scaling the service can be easily achieved.
In case there is a load balancer that handle this transparently, this error may not be received.
*/
type WebsocketServiceUnavailable struct {
	Payload *models.ErrorResponseBody
}

func (o *WebsocketServiceUnavailable) Error() string {
	return fmt.Sprintf("[GET /ws][%d] websocketServiceUnavailable  %+v", 503, o.Payload)
}

func (o *WebsocketServiceUnavailable) GetPayload() *models.ErrorResponseBody {
	return o.Payload
}

func (o *WebsocketServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponseBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
