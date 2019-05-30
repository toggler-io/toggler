package httpapi_test

import (
	"bytes"
	"fmt"
	"github.com/adamluzsi/FeatureFlags/extintf/http/httpapi"
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func NewServeMux(t *testcase.T) *httpapi.ServeMux {
	return httpapi.NewServeMux(usecases.NewUseCases(GetStorage(t)))
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
