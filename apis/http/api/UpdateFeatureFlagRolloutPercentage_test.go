package api_test

import (
	"bytes"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	. "github.com/adamluzsi/FeatureFlags/testing"
)

func TestServeMux_UpdateFeatureFlagRolloutPercentage(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	subject := func(t *testcase.T) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		NewServeMux(t).ServeHTTP(rr, t.I(`request`).(*http.Request))
		return rr
	}

	SetupSpecCommonVariables(s)

	s.Let(`percentage query value`, func(t *testcase.T) interface{} {
		return strconv.Itoa(GetRolloutPercentage(t))
	})

	s.Let(`request`, func(t *testcase.T) interface{} {
		u, err := url.Parse(`/update-feature-flag-rollout-percentage.json`)
		require.Nil(t, err)

		values := u.Query()
		values.Set(`token`, t.I(`TokenString`).(string))
		values.Set(`feature`, GetFeatureFlagName(t))
		values.Set(`percentage`, t.I(`percentage query value`).(string))
		u.RawQuery = values.Encode()

		return httptest.NewRequest(http.MethodGet, u.String(), bytes.NewBuffer([]byte{}))
	})

	s.When(`invalid percentage given`, func(s *testcase.Spec) {
		s.Let(`TokenString`, func(t *testcase.T) interface{} { return CreateToken(t, `manager who send bad requests`).Token })

		s.And(`as not a number`, func(s *testcase.Spec) {
			s.Let(`percentage query value`, func(t *testcase.T) interface{} {
				return `invalid`
			})

			s.Then(`it will return bad request`, func(t *testcase.T) {
				r := subject(t)

				require.Equal(t, 400, r.Code)
			})
		})

		s.And(`as not a number`, func(s *testcase.Spec) {
			s.Let(`percentage query value`, func(t *testcase.T) interface{} {
				return `-1`
			})

			s.Then(`it will return bad request`, func(t *testcase.T) {
				r := subject(t)

				require.Equal(t, 400, r.Code)
			})
		})
	})

	s.When(`invalid token given`, func(s *testcase.Spec) {
		s.Let(`TokenString`, func(t *testcase.T) interface{} {
			return `invalid`
		})

		s.Then(`it will return unauthorized`, func(t *testcase.T) {
			r := subject(t)

			require.Equal(t, 401, r.Code)
		})
	})

	s.When(`valid token provided`, func(s *testcase.Spec) {
		s.Let(`TokenString`, func(t *testcase.T) interface{} {
			return CreateSecurityTokenString(t)
		})

		s.Then(`call succeed`, func(t *testcase.T) {
			r := subject(t)
			require.Equal(t, 200, r.Code)
		})

		s.Then(`flag rollout percentage updated`, func(t *testcase.T) {
			r := subject(t)
			require.Equal(t, 200, r.Code)

			var resp struct{}
			IsJsonRespone(t, r, &resp)
			
			flag, err := GetStorage(t).FindFlagByName(GetFeatureFlagName(t))
			require.Nil(t, err)
			require.NotNil(t, flag)
			require.Equal(t, GetRolloutPercentage(t), flag.Rollout.Strategy.Percentage)
		})
	})

}
