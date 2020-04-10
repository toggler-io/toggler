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

// IsFeatureGloballyEnabledReader is a Reader for the IsFeatureGloballyEnabled structure.
type IsFeatureGloballyEnabledReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *IsFeatureGloballyEnabledReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewIsFeatureGloballyEnabledOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewIsFeatureGloballyEnabledBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewIsFeatureGloballyEnabledInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewIsFeatureGloballyEnabledOK creates a IsFeatureGloballyEnabledOK with default headers values
func NewIsFeatureGloballyEnabledOK() *IsFeatureGloballyEnabledOK {
	return &IsFeatureGloballyEnabledOK{}
}

/*IsFeatureGloballyEnabledOK handles this case with default header values.

EnrollmentResponse returns information about the requester's rollout feature enrollment status.
*/
type IsFeatureGloballyEnabledOK struct {
	Payload *models.EnrollmentResponseBody
}

func (o *IsFeatureGloballyEnabledOK) Error() string {
	return fmt.Sprintf("[POST /release/is-feature-globally-enabled.json][%d] isFeatureGloballyEnabledOK  %+v", 200, o.Payload)
}

func (o *IsFeatureGloballyEnabledOK) GetPayload() *models.EnrollmentResponseBody {
	return o.Payload
}

func (o *IsFeatureGloballyEnabledOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.EnrollmentResponseBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewIsFeatureGloballyEnabledBadRequest creates a IsFeatureGloballyEnabledBadRequest with default headers values
func NewIsFeatureGloballyEnabledBadRequest() *IsFeatureGloballyEnabledBadRequest {
	return &IsFeatureGloballyEnabledBadRequest{}
}

/*IsFeatureGloballyEnabledBadRequest handles this case with default header values.

ErrorResponse will contains a response about request that had some kind of problem.
The details will be included in the body.
*/
type IsFeatureGloballyEnabledBadRequest struct {
	Payload *models.ErrorResponseBody
}

func (o *IsFeatureGloballyEnabledBadRequest) Error() string {
	return fmt.Sprintf("[POST /release/is-feature-globally-enabled.json][%d] isFeatureGloballyEnabledBadRequest  %+v", 400, o.Payload)
}

func (o *IsFeatureGloballyEnabledBadRequest) GetPayload() *models.ErrorResponseBody {
	return o.Payload
}

func (o *IsFeatureGloballyEnabledBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponseBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewIsFeatureGloballyEnabledInternalServerError creates a IsFeatureGloballyEnabledInternalServerError with default headers values
func NewIsFeatureGloballyEnabledInternalServerError() *IsFeatureGloballyEnabledInternalServerError {
	return &IsFeatureGloballyEnabledInternalServerError{}
}

/*IsFeatureGloballyEnabledInternalServerError handles this case with default header values.

ErrorResponse will contains a response about request that had some kind of problem.
The details will be included in the body.
*/
type IsFeatureGloballyEnabledInternalServerError struct {
	Payload *models.ErrorResponseBody
}

func (o *IsFeatureGloballyEnabledInternalServerError) Error() string {
	return fmt.Sprintf("[POST /release/is-feature-globally-enabled.json][%d] isFeatureGloballyEnabledInternalServerError  %+v", 500, o.Payload)
}

func (o *IsFeatureGloballyEnabledInternalServerError) GetPayload() *models.ErrorResponseBody {
	return o.Payload
}

func (o *IsFeatureGloballyEnabledInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponseBody)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
