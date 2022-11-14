// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewStopFirehoseParams creates a new StopFirehoseParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewStopFirehoseParams() *StopFirehoseParams {
	return &StopFirehoseParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewStopFirehoseParamsWithTimeout creates a new StopFirehoseParams object
// with the ability to set a timeout on a request.
func NewStopFirehoseParamsWithTimeout(timeout time.Duration) *StopFirehoseParams {
	return &StopFirehoseParams{
		timeout: timeout,
	}
}

// NewStopFirehoseParamsWithContext creates a new StopFirehoseParams object
// with the ability to set a context for a request.
func NewStopFirehoseParamsWithContext(ctx context.Context) *StopFirehoseParams {
	return &StopFirehoseParams{
		Context: ctx,
	}
}

// NewStopFirehoseParamsWithHTTPClient creates a new StopFirehoseParams object
// with the ability to set a custom HTTPClient for a request.
func NewStopFirehoseParamsWithHTTPClient(client *http.Client) *StopFirehoseParams {
	return &StopFirehoseParams{
		HTTPClient: client,
	}
}

/*
StopFirehoseParams contains all the parameters to send to the API endpoint

	for the stop firehose operation.

	Typically these are written to a http.Request.
*/
type StopFirehoseParams struct {

	// Body.
	Body interface{}

	/* FirehoseUrn.

	   URN of the firehose.
	*/
	FirehoseUrn string

	/* ProjectID.

	   Unique identifier of the project.
	*/
	ProjectID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the stop firehose params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *StopFirehoseParams) WithDefaults() *StopFirehoseParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the stop firehose params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *StopFirehoseParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the stop firehose params
func (o *StopFirehoseParams) WithTimeout(timeout time.Duration) *StopFirehoseParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the stop firehose params
func (o *StopFirehoseParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the stop firehose params
func (o *StopFirehoseParams) WithContext(ctx context.Context) *StopFirehoseParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the stop firehose params
func (o *StopFirehoseParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the stop firehose params
func (o *StopFirehoseParams) WithHTTPClient(client *http.Client) *StopFirehoseParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the stop firehose params
func (o *StopFirehoseParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the stop firehose params
func (o *StopFirehoseParams) WithBody(body interface{}) *StopFirehoseParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the stop firehose params
func (o *StopFirehoseParams) SetBody(body interface{}) {
	o.Body = body
}

// WithFirehoseUrn adds the firehoseUrn to the stop firehose params
func (o *StopFirehoseParams) WithFirehoseUrn(firehoseUrn string) *StopFirehoseParams {
	o.SetFirehoseUrn(firehoseUrn)
	return o
}

// SetFirehoseUrn adds the firehoseUrn to the stop firehose params
func (o *StopFirehoseParams) SetFirehoseUrn(firehoseUrn string) {
	o.FirehoseUrn = firehoseUrn
}

// WithProjectID adds the projectID to the stop firehose params
func (o *StopFirehoseParams) WithProjectID(projectID string) *StopFirehoseParams {
	o.SetProjectID(projectID)
	return o
}

// SetProjectID adds the projectId to the stop firehose params
func (o *StopFirehoseParams) SetProjectID(projectID string) {
	o.ProjectID = projectID
}

// WriteToRequest writes these params to a swagger request
func (o *StopFirehoseParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	// path param firehoseUrn
	if err := r.SetPathParam("firehoseUrn", o.FirehoseUrn); err != nil {
		return err
	}

	// path param projectId
	if err := r.SetPathParam("projectId", o.ProjectID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
