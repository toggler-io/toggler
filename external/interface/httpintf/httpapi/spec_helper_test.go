package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"

	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/require"

	. "github.com/toggler-io/toggler/testing"
)

func CreateSecurityTokenString(t *testcase.T) string {
	textToken, token, err := security.NewIssuer(ExampleStorage(t)).CreateNewToken(context.TODO(), ExampleUniqueUserID(t), nil, nil)
	require.Nil(t, err)
	require.NotNil(t, token)
	require.NotEmpty(t, token.SHA512)
	return textToken
}

func IsJsonResponse(t *testcase.T, r *httptest.ResponseRecorder, ptr interface{}) {
	require.Equal(t, "application/json", r.Header().Get(`Content-Type`))
	require.Nil(t, json.NewDecoder(r.Body).Decode(ptr))
}

func publicAuth(t *testcase.T) runtime.ClientAuthInfoWriterFunc {
	return func(request runtime.ClientRequest, registry strfmt.Registry) error {
		return request.SetHeaderParam(`X-APP-KEY`, `example-app-key`)
	}
}

func protectedAuth(t *testcase.T) runtime.ClientAuthInfoWriterFunc {
	return func(request runtime.ClientRequest, registry strfmt.Registry) error {
		if err := publicAuth(t)(request, registry); err != nil {
			return err
		}
		
		return request.SetHeaderParam(`X-APP-TOKEN`, ExampleTextToken(t))
	}
}
