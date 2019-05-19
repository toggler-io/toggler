package api_test

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
		u, err := url.Parse(`/is-feature-globally-enabled`)
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
			rr := subject(t)

			require.Equal(t, 200, rr.Code)
		})
	})

	s.When(`flag is not global`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			UpdateFeatureFlagRolloutPercentage(t, GetFeatureFlagName(t), 50)
		})

		s.Then(`the request will be marked as forbidden`, func(t *testcase.T) {
			rr := subject(t)

			require.Equal(t, 403, rr.Code)
		})
	})

}

