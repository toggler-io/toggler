package api_test

import (
	"bytes"
	"encoding/json"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/security"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/adamluzsi/FeatureFlags/testing"
)

func TestServeMux_Li(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	subject := func(t *testcase.T) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		NewServeMux(t).ServeHTTP(rr, t.I(`request`).(*http.Request))
		return rr
	}

	SetupSpecCommonVariables(s)

	s.Let(`request`, func(t *testcase.T) interface{} {
		u, err := url.Parse(`/list-feature-flags.json`)
		require.Nil(t, err)

		values := u.Query()
		values.Set(`token`, t.I(`TokenString`).(string))
		u.RawQuery = values.Encode()

		return httptest.NewRequest(http.MethodGet, u.String(), bytes.NewBuffer([]byte{}))
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

		s.And(`no flag present in the system`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, GetStorage(t).Truncate(security.Token{}))
			})

			s.Then(`empty result received`, func(t *testcase.T) {
				r := subject(t)
				require.Equal(t, 200, r.Code)
				require.Equal(t, "application/json", r.Header().Get(`Content-Type`))
				require.Contains(t, r.Body.String(), `[]`)
			})
		})

		s.And(`feature flag is present in the system`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				UpdateFeatureFlagRolloutPercentage(t, `a`, 10)
			})

			s.Then(`flags received back`, func(t *testcase.T) {
				r := subject(t)
				require.Equal(t, 200, r.Code)
				require.Equal(t, "application/json", r.Header().Get(`Content-Type`))

				decoder := json.NewDecoder(r.Body)
				var resps []rollouts.FeatureFlag
				require.Nil(t, decoder.Decode(&resps))
				require.Equal(t, 1, len(resps))
				require.Equal(t, `a`, resps[0].Name)
				require.Equal(t, 10, resps[0].Rollout.Strategy.Percentage)
			})

			s.Then(`flags received back with lowercase json key convention`, func(t *testcase.T) {
				r := subject(t)
				require.Contains(t, r.Body.String(), `"id":`)
				require.Contains(t, r.Body.String(), `"name":`)
				require.Contains(t, r.Body.String(), `"rollout":`)
				require.Contains(t, r.Body.String(), `"strategy":`)
				require.Contains(t, r.Body.String(), `"percentage":`)
				require.Contains(t, r.Body.String(), `"url":`)
				require.Contains(t, r.Body.String(), `"rand_seed_salt":`)
			})

			s.And(`even multiple flag in the system`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					UpdateFeatureFlagRolloutPercentage(t, `b`, 20)
					UpdateFeatureFlagRolloutPercentage(t, `c`, 30)
				})

				s.Then(`flags received back`, func(t *testcase.T) {
					r := subject(t)
					require.Equal(t, 200, r.Code)
					require.Equal(t, "application/json", r.Header().Get(`Content-Type`))

					decoder := json.NewDecoder(r.Body)

					var resps []map[string]interface{}
					require.Nil(t, decoder.Decode(&resps))
					require.Equal(t, 3, len(resps))
				})
			})
		})
	})

}
