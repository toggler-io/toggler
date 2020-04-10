// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// URL A URL represents a parsed URL (technically, a URI reference).
//
// The general form represented is:
//
// [scheme:][//[userinfo@]host][/]path[?query][#fragment]
//
// URLs that do not start with a slash after the scheme are interpreted as:
//
// scheme:opaque[?query][#fragment]
//
// Note that the Path field is stored in decoded form: /%47%6f%2f becomes /Go/.
// A consequence is that it is impossible to tell which slashes in the Path were
// slashes in the raw URL and which were %2f. This distinction is rarely important,
// but when it is, the code should use RawPath, an optional field which only gets
// set if the default encoding is different from Path.
//
// URL's String method uses the EscapedPath method to obtain the path. See the
// EscapedPath method for more details.
//
// swagger:model URL
type URL struct {

	// force query
	ForceQuery bool `json:"ForceQuery,omitempty"`

	// fragment
	Fragment string `json:"Fragment,omitempty"`

	// host
	Host string `json:"Host,omitempty"`

	// opaque
	Opaque string `json:"Opaque,omitempty"`

	// path
	Path string `json:"Path,omitempty"`

	// raw path
	RawPath string `json:"RawPath,omitempty"`

	// raw query
	RawQuery string `json:"RawQuery,omitempty"`

	// scheme
	Scheme string `json:"Scheme,omitempty"`

	// user
	User Userinfo `json:"User,omitempty"`
}

// Validate validates this URL
func (m *URL) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *URL) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *URL) UnmarshalBinary(b []byte) error {
	var res URL
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
