package httpapi_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/external/interface/httpintf/httpapi"
	. "github.com/toggler-io/toggler/testing"
	"github.com/toggler-io/toggler/usecases"
)

func NewHandler(t *testcase.T) *httpapi.Handler {
	return httpapi.NewHandler(usecases.NewUseCases(ExampleStorage(t)))
}

func TestServeMuxRoutingPOC(t *testing.T) {

	rootmux := http.NewServeMux()
	submux := http.NewServeMux()
	subsubmux := http.NewServeMux()

	submux.HandleFunc(`/handler`, func(w http.ResponseWriter, r *http.Request) {
		t.Log(r.URL.String())
		t.Log(r.URL.Path)
		_, err := fmt.Fprintln(w, `Hello, submux!`)
		require.Nil(t, err)
	})

	subsubmux.HandleFunc(`/handler`, func(w http.ResponseWriter, r *http.Request) {
		t.Log(r.URL.String())
		t.Log(r.URL.Path)
		_, err := fmt.Fprintln(w, `Hello, subsubmux!`)
		require.Nil(t, err)
	})

	rootmux.Handle(`/sub/`, http.StripPrefix(`/sub`, submux))
	submux.Handle(`/sub/`, http.StripPrefix(`/sub`, subsubmux))

	ping := func(urlPath string) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, urlPath, bytes.NewBuffer([]byte{}))
		require.Nil(t, err)

		rootmux.ServeHTTP(rr, req)
		t.Log(rr.Body.String())
		require.Equal(t, 200, rr.Code)
	}

	ping(`/sub/handler`)
	ping(`/sub/sub/handler`)

}
