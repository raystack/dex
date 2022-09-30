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

// NewGetFirehoseParams creates a new GetFirehoseParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetFirehoseParams() *GetFirehoseParams {
	return &GetFirehoseParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetFirehoseParamsWithTimeout creates a new GetFirehoseParams object
// with the ability to set a timeout on a request.
func NewGetFirehoseParamsWithTimeout(timeout time.Duration) *GetFirehoseParams {
	return &GetFirehoseParams{
		timeout: timeout,
	}
}

// NewGetFirehoseParamsWithContext creates a new GetFirehoseParams object
// with the ability to set a context for a request.
func NewGetFirehoseParamsWithContext(ctx context.Context) *GetFirehoseParams {
	return &GetFirehoseParams{
		Context: ctx,
	}
}

// NewGetFirehoseParamsWithHTTPClient creates a new GetFirehoseParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetFirehoseParamsWithHTTPClient(client *http.Client) *GetFirehoseParams {
	return &GetFirehoseParams{
		HTTPClient: client,
	}
}

/*
GetFirehoseParams contains all the parameters to send to the API endpoint

	for the get firehose operation.

	Typically these are written to a http.Request.
*/
type GetFirehoseParams struct {

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

// WithDefaults hydrates default values in the get firehose params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetFirehoseParams) WithDefaults() *GetFirehoseParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get firehose params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetFirehoseParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get firehose params
func (o *GetFirehoseParams) WithTimeout(timeout time.Duration) *GetFirehoseParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get firehose params
func (o *GetFirehoseParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get firehose params
func (o *GetFirehoseParams) WithContext(ctx context.Context) *GetFirehoseParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get firehose params
func (o *GetFirehoseParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get firehose params
func (o *GetFirehoseParams) WithHTTPClient(client *http.Client) *GetFirehoseParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get firehose params
func (o *GetFirehoseParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithFirehoseUrn adds the firehoseUrn to the get firehose params
func (o *GetFirehoseParams) WithFirehoseUrn(firehoseUrn string) *GetFirehoseParams {
	o.SetFirehoseUrn(firehoseUrn)
	return o
}

// SetFirehoseUrn adds the firehoseUrn to the get firehose params
func (o *GetFirehoseParams) SetFirehoseUrn(firehoseUrn string) {
	o.FirehoseUrn = firehoseUrn
}

// WithProjectID adds the projectID to the get firehose params
func (o *GetFirehoseParams) WithProjectID(projectID string) *GetFirehoseParams {
	o.SetProjectID(projectID)
	return o
}

// SetProjectID adds the projectId to the get firehose params
func (o *GetFirehoseParams) SetProjectID(projectID string) {
	o.ProjectID = projectID
}

// WriteToRequest writes these params to a swagger request
func (o *GetFirehoseParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

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