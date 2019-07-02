package httpapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/toggler/testing"
	"github.com/stretchr/testify/require"
)

func TestServeMux_IsFeatureGloballyEnabled(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	subject := func(t *testcase.T) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		NewServeMux(t).ServeHTTP(rr, t.I(`request`).(*http.Request))
		return rr
	}

	SetupSpecCommonVariables(s)

	s.Let(`request`, func(t *testcase.T) interface{} {
		u, err := url.Parse(`/feature/is-globally-enabled.json`)
		require.Nil(t, err)

		values := u.Query()
		values.Set(`feature`, GetFeatureFlagName(t))
		u.RawQuery = values.Encode()

		return httptest.NewRequest(http.MethodGet, u.String(), bytes.NewBuffer([]byte{}))
	})

	s.When(`flag global`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			UpdateFeatureFlagRolloutPercentage(t, GetFeatureFlagName(t), 100)
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

	s.When(`flag is not global`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			UpdateFeatureFlagRolloutPercentage(t, GetFeatureFlagName(t), 50)
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
