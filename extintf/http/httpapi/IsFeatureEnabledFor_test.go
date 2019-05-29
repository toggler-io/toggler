package httpapi_test

import (
	"bytes"
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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

		values := u.Query()
		values.Set(`feature`, GetFeatureFlagName(t))
		values.Set(`id`, GetExternalPilotID(t))
		u.RawQuery = values.Encode()

		return httptest.NewRequest(http.MethodGet, u.String(), bytes.NewBuffer([]byte{}))
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

			IsJsonRespone(t, r, &resp)
			require.Equal(t, true, resp.Enrollment)
		})
	})

	s.When(`pilot is not`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			SpecPilotEnrolmentIs(t, false)
		})

		s.Then(`the request will be marked as forbidden`, func(t *testcase.T) {
			r := subject(t)

			require.Equal(t, 403, r.Code)

			var resp struct {
				Enrollment bool `json:"enrollment"`
			}

			IsJsonRespone(t, r, &resp)
			require.Equal(t, false, resp.Enrollment)
		})
	})

}
