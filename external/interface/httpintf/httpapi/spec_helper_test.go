package httpapi_test

import (
	"encoding/json"
	"github.com/toggler-io/toggler/external/interface/httpintf"
	"net/http"
	"net/http/httptest"

	"github.com/adamluzsi/testcase"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/security"

	sh "github.com/toggler-io/toggler/spechelper"
)

func CreateSecurityTokenString(t *testcase.T) string {
	textToken, token, err := security.NewIssuer(sh.StorageGet(t)).CreateNewToken(sh.ContextGet(t), sh.ExampleUniqueUserID(t), nil, nil)
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

		return request.SetHeaderParam(`X-APP-TOKEN`, sh.ExampleTextToken(t))
	}
}

func NewHTTPServer(t *testcase.T) *httptest.Server {
	uc := sh.ExampleUseCases(t)
	mux, err := httpintf.NewServeMux(uc)
	require.Nil(t, err)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r.WithContext(sh.ContextGet(t)))
	})
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}
