// Code generated by go-swagger; DO NOT EDIT.

package release_flag

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewGetClientConfigParams creates a new GetClientConfigParams object
// with the default values initialized.
func NewGetClientConfigParams() *GetClientConfigParams {
	var ()
	return &GetClientConfigParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetClientConfigParamsWithTimeout creates a new GetClientConfigParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetClientConfigParamsWithTimeout(timeout time.Duration) *GetClientConfigParams {
	var ()
	return &GetClientConfigParams{

		timeout: timeout,
	}
}

// NewGetClientConfigParamsWithContext creates a new GetClientConfigParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetClientConfigParamsWithContext(ctx context.Context) *GetClientConfigParams {
	var ()
	return &GetClientConfigParams{

		Context: ctx,
	}
}

// NewGetClientConfigParamsWithHTTPClient creates a new GetClientConfigParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetClientConfigParamsWithHTTPClient(client *http.Client) *GetClientConfigParams {
	var ()
	return &GetClientConfigParams{
		HTTPClient: client,
	}
}

/*GetClientConfigParams contains all the parameters to send to the API endpoint
for the get client config operation typically these are written to a http.Request
*/
type GetClientConfigParams struct {

	/*Body*/
	Body GetClientConfigBody

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get client config params
func (o *GetClientConfigParams) WithTimeout(timeout time.Duration) *GetClientConfigParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get client config params
func (o *GetClientConfigParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get client config params
func (o *GetClientConfigParams) WithContext(ctx context.Context) *GetClientConfigParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get client config params
func (o *GetClientConfigParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get client config params
func (o *GetClientConfigParams) WithHTTPClient(client *http.Client) *GetClientConfigParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get client config params
func (o *GetClientConfigParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the get client config params
func (o *GetClientConfigParams) WithBody(body GetClientConfigBody) *GetClientConfigParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the get client config params
func (o *GetClientConfigParams) SetBody(body GetClientConfigBody) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *GetClientConfigParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if err := r.SetBodyParam(o.Body); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
