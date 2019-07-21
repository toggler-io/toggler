package httpapi

// EnrollmentResponse returns information about the requester's rollout feature enrollment status.
// swagger:response enrollmentResponse
type EnrollmentResponse struct {
	// in: body
	Body EnrollmentResponseBody
}

// EnrollmentResponse will be returned when feature flag status is requested.
// The content will be always given, regardless if the flag exists or not.
// This helps the developers to use it as a null object, regardless the toggler service state.
type EnrollmentResponseBody struct {
	// Rollout feature flag enrollment status.
	Enrollment bool `json:"enrollment"`
}
