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

// NewListFirehosesParams creates a new ListFirehosesParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListFirehosesParams() *ListFirehosesParams {
	return &ListFirehosesParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListFirehosesParamsWithTimeout creates a new ListFirehosesParams object
// with the ability to set a timeout on a request.
func NewListFirehosesParamsWithTimeout(timeout time.Duration) *ListFirehosesParams {
	return &ListFirehosesParams{
		timeout: timeout,
	}
}

// NewListFirehosesParamsWithContext creates a new ListFirehosesParams object
// with the ability to set a context for a request.
func NewListFirehosesParamsWithContext(ctx context.Context) *ListFirehosesParams {
	return &ListFirehosesParams{
		Context: ctx,
	}
}

// NewListFirehosesParamsWithHTTPClient creates a new ListFirehosesParams object
// with the ability to set a custom HTTPClient for a request.
func NewListFirehosesParamsWithHTTPClient(client *http.Client) *ListFirehosesParams {
	return &ListFirehosesParams{
		HTTPClient: client,
	}
}

/*
ListFirehosesParams contains all the parameters to send to the API endpoint

	for the list firehoses operation.

	Typically these are written to a http.Request.
*/
type ListFirehosesParams struct {

	/* Cluster.

	   Return firehoses belonging to only this cluster.
	*/
	Cluster *string

	/* ProjectID.

	   Unique identifier of the project.
	*/
	ProjectID string

	/* SinkType.

	   Return firehoses with this sink type.
	*/
	SinkType *string

	/* Status.

	   Return firehoses only with this status.
	*/
	Status *string

	/* StreamName.

	     Return firehoses that are consuming from this stream.
	Usually stream refers to the kafka cluster.

	*/
	StreamName *string

	/* Team.

	   Return firehoses belonging to only this team.
	*/
	Team *string

	/* TopicName.

	   Return firehoses that are consuming from this topic.
	*/
	TopicName *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list firehoses params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListFirehosesParams) WithDefaults() *ListFirehosesParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list firehoses params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListFirehosesParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list firehoses params
func (o *ListFirehosesParams) WithTimeout(timeout time.Duration) *ListFirehosesParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list firehoses params
func (o *ListFirehosesParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list firehoses params
func (o *ListFirehosesParams) WithContext(ctx context.Context) *ListFirehosesParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list firehoses params
func (o *ListFirehosesParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list firehoses params
func (o *ListFirehosesParams) WithHTTPClient(client *http.Client) *ListFirehosesParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list firehoses params
func (o *ListFirehosesParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithCluster adds the cluster to the list firehoses params
func (o *ListFirehosesParams) WithCluster(cluster *string) *ListFirehosesParams {
	o.SetCluster(cluster)
	return o
}

// SetCluster adds the cluster to the list firehoses params
func (o *ListFirehosesParams) SetCluster(cluster *string) {
	o.Cluster = cluster
}

// WithProjectID adds the projectID to the list firehoses params
func (o *ListFirehosesParams) WithProjectID(projectID string) *ListFirehosesParams {
	o.SetProjectID(projectID)
	return o
}

// SetProjectID adds the projectId to the list firehoses params
func (o *ListFirehosesParams) SetProjectID(projectID string) {
	o.ProjectID = projectID
}

// WithSinkType adds the sinkType to the list firehoses params
func (o *ListFirehosesParams) WithSinkType(sinkType *string) *ListFirehosesParams {
	o.SetSinkType(sinkType)
	return o
}

// SetSinkType adds the sinkType to the list firehoses params
func (o *ListFirehosesParams) SetSinkType(sinkType *string) {
	o.SinkType = sinkType
}

// WithStatus adds the status to the list firehoses params
func (o *ListFirehosesParams) WithStatus(status *string) *ListFirehosesParams {
	o.SetStatus(status)
	return o
}

// SetStatus adds the status to the list firehoses params
func (o *ListFirehosesParams) SetStatus(status *string) {
	o.Status = status
}

// WithStreamName adds the streamName to the list firehoses params
func (o *ListFirehosesParams) WithStreamName(streamName *string) *ListFirehosesParams {
	o.SetStreamName(streamName)
	return o
}

// SetStreamName adds the streamName to the list firehoses params
func (o *ListFirehosesParams) SetStreamName(streamName *string) {
	o.StreamName = streamName
}

// WithTeam adds the team to the list firehoses params
func (o *ListFirehosesParams) WithTeam(team *string) *ListFirehosesParams {
	o.SetTeam(team)
	return o
}

// SetTeam adds the team to the list firehoses params
func (o *ListFirehosesParams) SetTeam(team *string) {
	o.Team = team
}

// WithTopicName adds the topicName to the list firehoses params
func (o *ListFirehosesParams) WithTopicName(topicName *string) *ListFirehosesParams {
	o.SetTopicName(topicName)
	return o
}

// SetTopicName adds the topicName to the list firehoses params
func (o *ListFirehosesParams) SetTopicName(topicName *string) {
	o.TopicName = topicName
}

// WriteToRequest writes these params to a swagger request
func (o *ListFirehosesParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Cluster != nil {

		// query param cluster
		var qrCluster string

		if o.Cluster != nil {
			qrCluster = *o.Cluster
		}
		qCluster := qrCluster
		if qCluster != "" {

			if err := r.SetQueryParam("cluster", qCluster); err != nil {
				return err
			}
		}
	}

	// path param projectId
	if err := r.SetPathParam("projectId", o.ProjectID); err != nil {
		return err
	}

	if o.SinkType != nil {

		// query param sink_type
		var qrSinkType string

		if o.SinkType != nil {
			qrSinkType = *o.SinkType
		}
		qSinkType := qrSinkType
		if qSinkType != "" {

			if err := r.SetQueryParam("sink_type", qSinkType); err != nil {
				return err
			}
		}
	}

	if o.Status != nil {

		// query param status
		var qrStatus string

		if o.Status != nil {
			qrStatus = *o.Status
		}
		qStatus := qrStatus
		if qStatus != "" {

			if err := r.SetQueryParam("status", qStatus); err != nil {
				return err
			}
		}
	}

	if o.StreamName != nil {

		// query param stream_name
		var qrStreamName string

		if o.StreamName != nil {
			qrStreamName = *o.StreamName
		}
		qStreamName := qrStreamName
		if qStreamName != "" {

			if err := r.SetQueryParam("stream_name", qStreamName); err != nil {
				return err
			}
		}
	}

	if o.Team != nil {

		// query param team
		var qrTeam string

		if o.Team != nil {
			qrTeam = *o.Team
		}
		qTeam := qrTeam
		if qTeam != "" {

			if err := r.SetQueryParam("team", qTeam); err != nil {
				return err
			}
		}
	}

	if o.TopicName != nil {

		// query param topic_name
		var qrTopicName string

		if o.TopicName != nil {
			qrTopicName = *o.TopicName
		}
		qTopicName := qrTopicName
		if qTopicName != "" {

			if err := r.SetQueryParam("topic_name", qTopicName); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
