// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new operations API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for operations API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption is the option for Client methods
type ClientOption func(*runtime.ClientOperation)

// ClientService is the interface for Client methods
type ClientService interface {
	CreateFirehose(params *CreateFirehoseParams, opts ...ClientOption) (*CreateFirehoseCreated, error)

	GetFirehose(params *GetFirehoseParams, opts ...ClientOption) (*GetFirehoseOK, error)

	GetProjectByID(params *GetProjectByIDParams, opts ...ClientOption) (*GetProjectByIDOK, error)

	ListFirehoses(params *ListFirehosesParams, opts ...ClientOption) (*ListFirehosesOK, error)

	ListProjects(params *ListProjectsParams, opts ...ClientOption) (*ListProjectsOK, error)

	UpdateFirehose(params *UpdateFirehoseParams, opts ...ClientOption) (*UpdateFirehoseOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
CreateFirehose creates a new firehose

Create and deploy a new firehose as per the configurations in the body.
*/
func (a *Client) CreateFirehose(params *CreateFirehoseParams, opts ...ClientOption) (*CreateFirehoseCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateFirehoseParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "createFirehose",
		Method:             "POST",
		PathPattern:        "/projects/{projectId}/firehoses",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CreateFirehoseReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateFirehoseCreated)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for createFirehose: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetFirehose gets firehose by u r n

Get firehose by URN.
*/
func (a *Client) GetFirehose(params *GetFirehoseParams, opts ...ClientOption) (*GetFirehoseOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetFirehoseParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getFirehose",
		Method:             "GET",
		PathPattern:        "/projects/{projectId}/firehoses/{firehoseUrn}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetFirehoseReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetFirehoseOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getFirehose: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetProjectByID gets project by id

Get project by its unique identifier.
*/
func (a *Client) GetProjectByID(params *GetProjectByIDParams, opts ...ClientOption) (*GetProjectByIDOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetProjectByIDParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getProjectById",
		Method:             "GET",
		PathPattern:        "/projects/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetProjectByIDReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetProjectByIDOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getProjectById: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
ListFirehoses gets list of firehoses

Get list of firehoses in this project.
*/
func (a *Client) ListFirehoses(params *ListFirehosesParams, opts ...ClientOption) (*ListFirehosesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListFirehosesParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "listFirehoses",
		Method:             "GET",
		PathPattern:        "/projects/{projectId}/firehoses",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ListFirehosesReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ListFirehosesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for listFirehoses: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
ListProjects gets list of projects

Get list of projects.
*/
func (a *Client) ListProjects(params *ListProjectsParams, opts ...ClientOption) (*ListProjectsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListProjectsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "listProjects",
		Method:             "GET",
		PathPattern:        "/projects",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ListProjectsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ListProjectsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for listProjects: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
UpdateFirehose updates firehose configurations

Update firehose configurations.
*/
func (a *Client) UpdateFirehose(params *UpdateFirehoseParams, opts ...ClientOption) (*UpdateFirehoseOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewUpdateFirehoseParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "updateFirehose",
		Method:             "PUT",
		PathPattern:        "/projects/{projectId}/firehoses/{firehoseUrn}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &UpdateFirehoseReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*UpdateFirehoseOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for updateFirehose: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
