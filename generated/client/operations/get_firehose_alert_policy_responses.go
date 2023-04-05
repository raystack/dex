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

// GetFirehoseAlertPolicyReader is a Reader for the GetFirehoseAlertPolicy structure.
type GetFirehoseAlertPolicyReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetFirehoseAlertPolicyReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetFirehoseAlertPolicyOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewGetFirehoseAlertPolicyNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetFirehoseAlertPolicyInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetFirehoseAlertPolicyOK creates a GetFirehoseAlertPolicyOK with default headers values
func NewGetFirehoseAlertPolicyOK() *GetFirehoseAlertPolicyOK {
	return &GetFirehoseAlertPolicyOK{}
}

/*
GetFirehoseAlertPolicyOK describes a response with status code 200, with default header values.

Found alert policy for given firehose URN.
*/
type GetFirehoseAlertPolicyOK struct {
	Payload *models.AlertPolicy
}

// IsSuccess returns true when this get firehose alert policy o k response has a 2xx status code
func (o *GetFirehoseAlertPolicyOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get firehose alert policy o k response has a 3xx status code
func (o *GetFirehoseAlertPolicyOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get firehose alert policy o k response has a 4xx status code
func (o *GetFirehoseAlertPolicyOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get firehose alert policy o k response has a 5xx status code
func (o *GetFirehoseAlertPolicyOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get firehose alert policy o k response a status code equal to that given
func (o *GetFirehoseAlertPolicyOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetFirehoseAlertPolicyOK) Error() string {
	return fmt.Sprintf("[GET /dex/firehoses/{firehoseUrn}/alertPolicy][%d] getFirehoseAlertPolicyOK  %+v", 200, o.Payload)
}

func (o *GetFirehoseAlertPolicyOK) String() string {
	return fmt.Sprintf("[GET /dex/firehoses/{firehoseUrn}/alertPolicy][%d] getFirehoseAlertPolicyOK  %+v", 200, o.Payload)
}

func (o *GetFirehoseAlertPolicyOK) GetPayload() *models.AlertPolicy {
	return o.Payload
}

func (o *GetFirehoseAlertPolicyOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AlertPolicy)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFirehoseAlertPolicyNotFound creates a GetFirehoseAlertPolicyNotFound with default headers values
func NewGetFirehoseAlertPolicyNotFound() *GetFirehoseAlertPolicyNotFound {
	return &GetFirehoseAlertPolicyNotFound{}
}

/*
GetFirehoseAlertPolicyNotFound describes a response with status code 404, with default header values.

Firehose with given URN was not found
*/
type GetFirehoseAlertPolicyNotFound struct {
	Payload *models.ErrorResponse
}

// IsSuccess returns true when this get firehose alert policy not found response has a 2xx status code
func (o *GetFirehoseAlertPolicyNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get firehose alert policy not found response has a 3xx status code
func (o *GetFirehoseAlertPolicyNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get firehose alert policy not found response has a 4xx status code
func (o *GetFirehoseAlertPolicyNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get firehose alert policy not found response has a 5xx status code
func (o *GetFirehoseAlertPolicyNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get firehose alert policy not found response a status code equal to that given
func (o *GetFirehoseAlertPolicyNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetFirehoseAlertPolicyNotFound) Error() string {
	return fmt.Sprintf("[GET /dex/firehoses/{firehoseUrn}/alertPolicy][%d] getFirehoseAlertPolicyNotFound  %+v", 404, o.Payload)
}

func (o *GetFirehoseAlertPolicyNotFound) String() string {
	return fmt.Sprintf("[GET /dex/firehoses/{firehoseUrn}/alertPolicy][%d] getFirehoseAlertPolicyNotFound  %+v", 404, o.Payload)
}

func (o *GetFirehoseAlertPolicyNotFound) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetFirehoseAlertPolicyNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFirehoseAlertPolicyInternalServerError creates a GetFirehoseAlertPolicyInternalServerError with default headers values
func NewGetFirehoseAlertPolicyInternalServerError() *GetFirehoseAlertPolicyInternalServerError {
	return &GetFirehoseAlertPolicyInternalServerError{}
}

/*
GetFirehoseAlertPolicyInternalServerError describes a response with status code 500, with default header values.

internal error
*/
type GetFirehoseAlertPolicyInternalServerError struct {
	Payload *models.ErrorResponse
}

// IsSuccess returns true when this get firehose alert policy internal server error response has a 2xx status code
func (o *GetFirehoseAlertPolicyInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get firehose alert policy internal server error response has a 3xx status code
func (o *GetFirehoseAlertPolicyInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get firehose alert policy internal server error response has a 4xx status code
func (o *GetFirehoseAlertPolicyInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get firehose alert policy internal server error response has a 5xx status code
func (o *GetFirehoseAlertPolicyInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get firehose alert policy internal server error response a status code equal to that given
func (o *GetFirehoseAlertPolicyInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *GetFirehoseAlertPolicyInternalServerError) Error() string {
	return fmt.Sprintf("[GET /dex/firehoses/{firehoseUrn}/alertPolicy][%d] getFirehoseAlertPolicyInternalServerError  %+v", 500, o.Payload)
}

func (o *GetFirehoseAlertPolicyInternalServerError) String() string {
	return fmt.Sprintf("[GET /dex/firehoses/{firehoseUrn}/alertPolicy][%d] getFirehoseAlertPolicyInternalServerError  %+v", 500, o.Payload)
}

func (o *GetFirehoseAlertPolicyInternalServerError) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetFirehoseAlertPolicyInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
