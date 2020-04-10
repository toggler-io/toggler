// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// FlagRolloutStrategy flag rollout strategy
//
// swagger:model FlagRolloutStrategy
type FlagRolloutStrategy struct {

	// Percentage allows you to define how many of your user base should be enrolled pseudo randomly.
	Percentage int64 `json:"percentage,omitempty"`

	// decision logic api
	DecisionLogicAPI *URL `json:"decision_logic_api,omitempty"`
}

// Validate validates this flag rollout strategy
func (m *FlagRolloutStrategy) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDecisionLogicAPI(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *FlagRolloutStrategy) validateDecisionLogicAPI(formats strfmt.Registry) error {

	if swag.IsZero(m.DecisionLogicAPI) { // not required
		return nil
	}

	if m.DecisionLogicAPI != nil {
		if err := m.DecisionLogicAPI.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("decision_logic_api")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *FlagRolloutStrategy) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *FlagRolloutStrategy) UnmarshalBinary(b []byte) error {
	var res FlagRolloutStrategy
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
