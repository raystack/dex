// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"fmt"
	"io"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"

	"github.com/odpf/dex/generated/models"
)

// ScaleFirehoseReader is a Reader for the ScaleFirehose structure.
type ScaleFirehoseReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ScaleFirehoseReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewScaleFirehoseOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewScaleFirehoseBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewScaleFirehoseNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewScaleFirehoseInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewScaleFirehoseOK creates a ScaleFirehoseOK with default headers values
func NewScaleFirehoseOK() *ScaleFirehoseOK {
	return &ScaleFirehoseOK{}
}

/*
ScaleFirehoseOK describes a response with status code 200, with default header values.

Successfully applied update.
*/
type ScaleFirehoseOK struct {
	Payload *models.Firehose
}

// IsSuccess returns true when this scale firehose o k response has a 2xx status code
func (o *ScaleFirehoseOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this scale firehose o k response has a 3xx status code
func (o *ScaleFirehoseOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this scale firehose o k response has a 4xx status code
func (o *ScaleFirehoseOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this scale firehose o k response has a 5xx status code
func (o *ScaleFirehoseOK) IsServerError() bool {
	return false
}

// IsCode returns true when this scale firehose o k response a status code equal to that given
func (o *ScaleFirehoseOK) IsCode(code int) bool {
	return code == 200
}

func (o *ScaleFirehoseOK) Error() string {
	return fmt.Sprintf("[POST /projects/{projectSlug}/firehoses/{firehoseUrn}/scale][%d] scaleFirehoseOK  %+v", 200, o.Payload)
}

func (o *ScaleFirehoseOK) String() string {
	return fmt.Sprintf("[POST /projects/{projectSlug}/firehoses/{firehoseUrn}/scale][%d] scaleFirehoseOK  %+v", 200, o.Payload)
}

func (o *ScaleFirehoseOK) GetPayload() *models.Firehose {
	return o.Payload
}

func (o *ScaleFirehoseOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Firehose)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewScaleFirehoseBadRequest creates a ScaleFirehoseBadRequest with default headers values
func NewScaleFirehoseBadRequest() *ScaleFirehoseBadRequest {
	return &ScaleFirehoseBadRequest{}
}

/*
ScaleFirehoseBadRequest describes a response with status code 400, with default header values.

Update request is not valid.
*/
type ScaleFirehoseBadRequest struct {
	Payload *models.ErrorResponse
}

// IsSuccess returns true when this scale firehose bad request response has a 2xx status code
func (o *ScaleFirehoseBadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this scale firehose bad request response has a 3xx status code
func (o *ScaleFirehoseBadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this scale firehose bad request response has a 4xx status code
func (o *ScaleFirehoseBadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this scale firehose bad request response has a 5xx status code
func (o *ScaleFirehoseBadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this scale firehose bad request response a status code equal to that given
func (o *ScaleFirehoseBadRequest) IsCode(code int) bool {
	return code == 400
}

func (o *ScaleFirehoseBadRequest) Error() string {
	return fmt.Sprintf("[POST /projects/{projectSlug}/firehoses/{firehoseUrn}/scale][%d] scaleFirehoseBadRequest  %+v", 400, o.Payload)
}

func (o *ScaleFirehoseBadRequest) String() string {
	return fmt.Sprintf("[POST /projects/{projectSlug}/firehoses/{firehoseUrn}/scale][%d] scaleFirehoseBadRequest  %+v", 400, o.Payload)
}

func (o *ScaleFirehoseBadRequest) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *ScaleFirehoseBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewScaleFirehoseNotFound creates a ScaleFirehoseNotFound with default headers values
func NewScaleFirehoseNotFound() *ScaleFirehoseNotFound {
	return &ScaleFirehoseNotFound{}
}

/*
ScaleFirehoseNotFound describes a response with status code 404, with default header values.

Firehose with given URN was not found
*/
type ScaleFirehoseNotFound struct {
	Payload *models.ErrorResponse
}

// IsSuccess returns true when this scale firehose not found response has a 2xx status code
func (o *ScaleFirehoseNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this scale firehose not found response has a 3xx status code
func (o *ScaleFirehoseNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this scale firehose not found response has a 4xx status code
func (o *ScaleFirehoseNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this scale firehose not found response has a 5xx status code
func (o *ScaleFirehoseNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this scale firehose not found response a status code equal to that given
func (o *ScaleFirehoseNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *ScaleFirehoseNotFound) Error() string {
	return fmt.Sprintf("[POST /projects/{projectSlug}/firehoses/{firehoseUrn}/scale][%d] scaleFirehoseNotFound  %+v", 404, o.Payload)
}

func (o *ScaleFirehoseNotFound) String() string {
	return fmt.Sprintf("[POST /projects/{projectSlug}/firehoses/{firehoseUrn}/scale][%d] scaleFirehoseNotFound  %+v", 404, o.Payload)
}

func (o *ScaleFirehoseNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *ScaleFirehoseNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewScaleFirehoseInternalServerError creates a ScaleFirehoseInternalServerError with default headers values
func NewScaleFirehoseInternalServerError() *ScaleFirehoseInternalServerError {
	return &ScaleFirehoseInternalServerError{}
}

/*
ScaleFirehoseInternalServerError describes a response with status code 500, with default header values.

internal error
*/
type ScaleFirehoseInternalServerError struct {
	Payload *models.ErrorResponse
}

// IsSuccess returns true when this scale firehose internal server error response has a 2xx status code
func (o *ScaleFirehoseInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this scale firehose internal server error response has a 3xx status code
func (o *ScaleFirehoseInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this scale firehose internal server error response has a 4xx status code
func (o *ScaleFirehoseInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this scale firehose internal server error response has a 5xx status code
func (o *ScaleFirehoseInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this scale firehose internal server error response a status code equal to that given
func (o *ScaleFirehoseInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *ScaleFirehoseInternalServerError) Error() string {
	return fmt.Sprintf("[POST /projects/{projectSlug}/firehoses/{firehoseUrn}/scale][%d] scaleFirehoseInternalServerError  %+v", 500, o.Payload)
}

func (o *ScaleFirehoseInternalServerError) String() string {
	return fmt.Sprintf("[POST /projects/{projectSlug}/firehoses/{firehoseUrn}/scale][%d] scaleFirehoseInternalServerError  %+v", 500, o.Payload)
}

func (o *ScaleFirehoseInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *ScaleFirehoseInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

/*
ScaleFirehoseBody scale firehose body
swagger:model ScaleFirehoseBody
*/
type ScaleFirehoseBody struct {

	// Number of replicas to run.
	// Example: 2
	// Required: true
	Replicas *float64 `json:"replicas"`
}

// Validate validates this scale firehose body
func (o *ScaleFirehoseBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateReplicas(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *ScaleFirehoseBody) validateReplicas(formats strfmt.Registry) error {

	if err := validate.Required("body"+"."+"replicas", "body", o.Replicas); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this scale firehose body based on context it is used
func (o *ScaleFirehoseBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (o *ScaleFirehoseBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *ScaleFirehoseBody) UnmarshalBinary(b []byte) error {
	var res ScaleFirehoseBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
