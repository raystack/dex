// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// FirehosePartialConfig firehose partial config
//
// swagger:model FirehosePartialConfig
type FirehosePartialConfig struct {

	// If specified, it will be merged with the current configs.
	EnvVars map[string]string `json:"env_vars,omitempty"`

	// Set to non-empty string to update
	// Example: gotocompany/firehose:0.1.0
	Image string `json:"image,omitempty"`

	// Set to a value greater than 0 to update.
	Replicas float64 `json:"replicas,omitempty"`

	// - Omit this field to not update.
	// - Setting it to a non-empty string to update.
	// - Set it to empty string to remove the current stop_time.
	//
	StopTime *string `json:"stop_time,omitempty"`

	// Set to true/false to stop or start the firehose.
	Stopped *bool `json:"stopped,omitempty"`

	// Set to non-empty string to update
	StreamName string `json:"stream_name,omitempty"`
}

// Validate validates this firehose partial config
func (m *FirehosePartialConfig) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this firehose partial config based on context it is used
func (m *FirehosePartialConfig) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *FirehosePartialConfig) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *FirehosePartialConfig) UnmarshalBinary(b []byte) error {
	var res FirehosePartialConfig
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
