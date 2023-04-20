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
	"github.com/go-openapi/swag"
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

	/* Group.

	   Return firehoses belonging to only this group.
	*/
	Group *string

	/* KubeCluster.

	   Return firehoses belonging to only this kubernetes cluster.
	*/
	KubeCluster *string

	/* Project.

	   Unique identifier of the project.
	*/
	Project string

	/* SinkType.

	   Return firehoses with this sink type.
	*/
	SinkType []string

	/* Status.

	   Return firehoses only with this status.
	*/
	Status *string

	/* StreamName.

	     Return firehoses that are consuming from this stream.
	Usually stream refers to the kafka cluster.

	*/
	StreamName *string

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

// WithGroup adds the group to the list firehoses params
func (o *ListFirehosesParams) WithGroup(group *string) *ListFirehosesParams {
	o.SetGroup(group)
	return o
}

// SetGroup adds the group to the list firehoses params
func (o *ListFirehosesParams) SetGroup(group *string) {
	o.Group = group
}

// WithKubeCluster adds the kubeCluster to the list firehoses params
func (o *ListFirehosesParams) WithKubeCluster(kubeCluster *string) *ListFirehosesParams {
	o.SetKubeCluster(kubeCluster)
	return o
}

// SetKubeCluster adds the kubeCluster to the list firehoses params
func (o *ListFirehosesParams) SetKubeCluster(kubeCluster *string) {
	o.KubeCluster = kubeCluster
}

// WithProject adds the project to the list firehoses params
func (o *ListFirehosesParams) WithProject(project string) *ListFirehosesParams {
	o.SetProject(project)
	return o
}

// SetProject adds the project to the list firehoses params
func (o *ListFirehosesParams) SetProject(project string) {
	o.Project = project
}

// WithSinkType adds the sinkType to the list firehoses params
func (o *ListFirehosesParams) WithSinkType(sinkType []string) *ListFirehosesParams {
	o.SetSinkType(sinkType)
	return o
}

// SetSinkType adds the sinkType to the list firehoses params
func (o *ListFirehosesParams) SetSinkType(sinkType []string) {
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

	if o.Group != nil {

		// query param group
		var qrGroup string

		if o.Group != nil {
			qrGroup = *o.Group
		}
		qGroup := qrGroup
		if qGroup != "" {

			if err := r.SetQueryParam("group", qGroup); err != nil {
				return err
			}
		}
	}

	if o.KubeCluster != nil {

		// query param kube_cluster
		var qrKubeCluster string

		if o.KubeCluster != nil {
			qrKubeCluster = *o.KubeCluster
		}
		qKubeCluster := qrKubeCluster
		if qKubeCluster != "" {

			if err := r.SetQueryParam("kube_cluster", qKubeCluster); err != nil {
				return err
			}
		}
	}

	// query param project
	qrProject := o.Project
	qProject := qrProject
	if qProject != "" {

		if err := r.SetQueryParam("project", qProject); err != nil {
			return err
		}
	}

	if o.SinkType != nil {

		// binding items for sink_type
		joinedSinkType := o.bindParamSinkType(reg)

		// query array param sink_type
		if err := r.SetQueryParam("sink_type", joinedSinkType...); err != nil {
			return err
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

// bindParamListFirehoses binds the parameter sink_type
func (o *ListFirehosesParams) bindParamSinkType(formats strfmt.Registry) []string {
	sinkTypeIR := o.SinkType

	var sinkTypeIC []string
	for _, sinkTypeIIR := range sinkTypeIR { // explode []string

		sinkTypeIIV := sinkTypeIIR // string as string
		sinkTypeIC = append(sinkTypeIC, sinkTypeIIV)
	}

	// items.CollectionFormat: "csv"
	sinkTypeIS := swag.JoinByFormat(sinkTypeIC, "csv")

	return sinkTypeIS
}
