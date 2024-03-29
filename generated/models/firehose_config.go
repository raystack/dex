// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// FirehoseConfig firehose config
//
// swagger:model FirehoseConfig
type FirehoseConfig struct {

	// bootstrap servers
	// Required: true
	BootstrapServers *string `json:"bootstrap_servers"`

	// consumer group id
	// Required: true
	ConsumerGroupID *string `json:"consumer_group_id"`

	// env vars
	EnvVars map[string]string `json:"env_vars,omitempty"`

	// input schema proto class
	// Required: true
	InputSchemaProtoClass *string `json:"input_schema_proto_class"`

	// replicas
	Replicas *float64 `json:"replicas,omitempty"`

	// sink type
	// Required: true
	SinkType *FirehoseSinkType `json:"sink_type"`

	// stop date
	StopDate string `json:"stop_date,omitempty"`

	// stream name
	// Required: true
	StreamName *string `json:"stream_name"`

	// topic name
	// Required: true
	TopicName *string `json:"topic_name"`

	// version
	// Example: 1.0.0
	// Read Only: true
	Version string `json:"version,omitempty"`
}

// Validate validates this firehose config
func (m *FirehoseConfig) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateBootstrapServers(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateConsumerGroupID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateInputSchemaProtoClass(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSinkType(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStreamName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTopicName(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *FirehoseConfig) validateBootstrapServers(formats strfmt.Registry) error {

	if err := validate.Required("bootstrap_servers", "body", m.BootstrapServers); err != nil {
		return err
	}

	return nil
}

func (m *FirehoseConfig) validateConsumerGroupID(formats strfmt.Registry) error {

	if err := validate.Required("consumer_group_id", "body", m.ConsumerGroupID); err != nil {
		return err
	}

	return nil
}

func (m *FirehoseConfig) validateInputSchemaProtoClass(formats strfmt.Registry) error {

	if err := validate.Required("input_schema_proto_class", "body", m.InputSchemaProtoClass); err != nil {
		return err
	}

	return nil
}

func (m *FirehoseConfig) validateSinkType(formats strfmt.Registry) error {

	if err := validate.Required("sink_type", "body", m.SinkType); err != nil {
		return err
	}

	if err := validate.Required("sink_type", "body", m.SinkType); err != nil {
		return err
	}

	if m.SinkType != nil {
		if err := m.SinkType.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("sink_type")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("sink_type")
			}
			return err
		}
	}

	return nil
}

func (m *FirehoseConfig) validateStreamName(formats strfmt.Registry) error {

	if err := validate.Required("stream_name", "body", m.StreamName); err != nil {
		return err
	}

	return nil
}

func (m *FirehoseConfig) validateTopicName(formats strfmt.Registry) error {

	if err := validate.Required("topic_name", "body", m.TopicName); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this firehose config based on the context it is used
func (m *FirehoseConfig) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateSinkType(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateVersion(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *FirehoseConfig) contextValidateSinkType(ctx context.Context, formats strfmt.Registry) error {

	if m.SinkType != nil {
		if err := m.SinkType.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("sink_type")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("sink_type")
			}
			return err
		}
	}

	return nil
}

func (m *FirehoseConfig) contextValidateVersion(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.ReadOnly(ctx, "version", "body", string(m.Version)); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *FirehoseConfig) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *FirehoseConfig) UnmarshalBinary(b []byte) error {
	var res FirehoseConfig
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
