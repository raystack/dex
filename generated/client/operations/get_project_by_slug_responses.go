// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/goto/dex/generated/models"
)

// GetProjectBySlugReader is a Reader for the GetProjectBySlug structure.
type GetProjectBySlugReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetProjectBySlugReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetProjectBySlugOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewGetProjectBySlugNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetProjectBySlugInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetProjectBySlugOK creates a GetProjectBySlugOK with default headers values
func NewGetProjectBySlugOK() *GetProjectBySlugOK {
	return &GetProjectBySlugOK{}
}

/*
GetProjectBySlugOK describes a response with status code 200, with default header values.

successful operation
*/
type GetProjectBySlugOK struct {
	Payload *models.Project
}

// IsSuccess returns true when this get project by slug o k response has a 2xx status code
func (o *GetProjectBySlugOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get project by slug o k response has a 3xx status code
func (o *GetProjectBySlugOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get project by slug o k response has a 4xx status code
func (o *GetProjectBySlugOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get project by slug o k response has a 5xx status code
func (o *GetProjectBySlugOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get project by slug o k response a status code equal to that given
func (o *GetProjectBySlugOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetProjectBySlugOK) Error() string {
	return fmt.Sprintf("[GET /projects/{slug}][%d] getProjectBySlugOK  %+v", 200, o.Payload)
}

func (o *GetProjectBySlugOK) String() string {
	return fmt.Sprintf("[GET /projects/{slug}][%d] getProjectBySlugOK  %+v", 200, o.Payload)
}

func (o *GetProjectBySlugOK) GetPayload() *models.Project {
	return o.Payload
}

func (o *GetProjectBySlugOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Project)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetProjectBySlugNotFound creates a GetProjectBySlugNotFound with default headers values
func NewGetProjectBySlugNotFound() *GetProjectBySlugNotFound {
	return &GetProjectBySlugNotFound{}
}

/*
GetProjectBySlugNotFound describes a response with status code 404, with default header values.

project not found
*/
type GetProjectBySlugNotFound struct {
	Payload *models.ErrorResponse
}

// IsSuccess returns true when this get project by slug not found response has a 2xx status code
func (o *GetProjectBySlugNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get project by slug not found response has a 3xx status code
func (o *GetProjectBySlugNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get project by slug not found response has a 4xx status code
func (o *GetProjectBySlugNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get project by slug not found response has a 5xx status code
func (o *GetProjectBySlugNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get project by slug not found response a status code equal to that given
func (o *GetProjectBySlugNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetProjectBySlugNotFound) Error() string {
	return fmt.Sprintf("[GET /projects/{slug}][%d] getProjectBySlugNotFound  %+v", 404, o.Payload)
}

func (o *GetProjectBySlugNotFound) String() string {
	return fmt.Sprintf("[GET /projects/{slug}][%d] getProjectBySlugNotFound  %+v", 404, o.Payload)
}

func (o *GetProjectBySlugNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetProjectBySlugNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetProjectBySlugInternalServerError creates a GetProjectBySlugInternalServerError with default headers values
func NewGetProjectBySlugInternalServerError() *GetProjectBySlugInternalServerError {
	return &GetProjectBySlugInternalServerError{}
}

/*
GetProjectBySlugInternalServerError describes a response with status code 500, with default header values.

internal error
*/
type GetProjectBySlugInternalServerError struct {
	Payload *models.ErrorResponse
}

// IsSuccess returns true when this get project by slug internal server error response has a 2xx status code
func (o *GetProjectBySlugInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get project by slug internal server error response has a 3xx status code
func (o *GetProjectBySlugInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get project by slug internal server error response has a 4xx status code
func (o *GetProjectBySlugInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get project by slug internal server error response has a 5xx status code
func (o *GetProjectBySlugInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get project by slug internal server error response a status code equal to that given
func (o *GetProjectBySlugInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *GetProjectBySlugInternalServerError) Error() string {
	return fmt.Sprintf("[GET /projects/{slug}][%d] getProjectBySlugInternalServerError  %+v", 500, o.Payload)
}

func (o *GetProjectBySlugInternalServerError) String() string {
	return fmt.Sprintf("[GET /projects/{slug}][%d] getProjectBySlugInternalServerError  %+v", 500, o.Payload)
}

func (o *GetProjectBySlugInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetProjectBySlugInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
