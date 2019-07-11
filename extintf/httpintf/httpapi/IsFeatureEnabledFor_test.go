package httpapi_test

import (
	"bytes"
	"encoding/json"
	"github.com/adamluzsi/toggler/extintf/httpintf/httpapi"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/toggler/testing"
	"github.com/stretchr/testify/require"
)

func TestServeMux_IsFeatureEnabledFor(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	subject := func(t *testcase.T) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		NewServeMux(t).IsFeatureEnabledFor(rr, t.I(`request`).(*http.Request))
		return rr
	}

	SetupSpecCommonVariables(s)

	s.Let(`request`, func(t *testcase.T) interface{} {
		u, err := url.Parse(`/feature/is-enabled.json`)
		require.Nil(t, err)
		payload := bytes.NewBuffer([]byte{})
		jsonenc := json.NewEncoder(payload)
		require.Nil(t, jsonenc.Encode(httpapi.IsFeatureEnabledForReqBody{
			Feature: GetFeatureFlagName(t),
			PilotID: GetExternalPilotID(t),
		}))
		return httptest.NewRequest(http.MethodGet, u.String(), payload)
	})

	s.When(`pilot is enrolled`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			SpecPilotEnrolmentIs(t, true)
		})

		s.Then(`the request will be accepted with OK`, func(t *testcase.T) {
			r := subject(t)

			require.Equal(t, 200, r.Code)

			var resp struct {
				Enrollment bool `json:"enrollment"`
			}

			IsJsonResponse(t, r, &resp)
			require.Equal(t, true, resp.Enrollment)
		})
	})

	s.When(`pilot is not`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			SpecPilotEnrolmentIs(t, false)
		})

		s.Then(`the request will be marked as forbidden`, func(t *testcase.T) {
			r := subject(t)

			require.Equal(t, 200, r.Code)

			var resp struct {
				Enrollment bool `json:"enrollment"`
			}

			IsJsonResponse(t, r, &resp)
			require.Equal(t, false, resp.Enrollment)
		})
	})

}
