package httpapi_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/stretchr/testify/require"

	. "github.com/toggler-io/toggler/testing"
)

func TestServeMux_ListFeatureFlags(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	subject := func(t *testcase.T) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		NewHandler(t).ServeHTTP(rr, t.I(`request`).(*http.Request))
		return rr
	}

	SetupSpecCommonVariables(s)

	s.Let(`request`, func(t *testcase.T) interface{} {
		u, err := url.Parse(`/release/flag/list.json`)
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
				require.Nil(t, GetStorage(t).DeleteAll(context.Background(), security.Token{}))
			})

			s.Then(`empty result received`, func(t *testcase.T) {
				r := subject(t)
				require.Equal(t, 200, r.Code)

				require.Contains(t, r.Body.String(), `[]`)
				var flags []*release.Flag
				IsJsonResponse(t, r, &flags)
				require.Empty(t, flags)
			})
		})

		s.And(`feature flag is present in the system`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				UpdateReleaseFlagRolloutPercentage(t, `a`, 10)
			})

			s.Then(`flags received back`, func(t *testcase.T) {
				r := subject(t)
				require.Equal(t, 200, r.Code)

				var resps []release.Flag
				IsJsonResponse(t, r, &resps)

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
				require.Contains(t, r.Body.String(), `"decision_logic_api":`)
				require.Contains(t, r.Body.String(), `"rand_seed_salt":`)
			})

			s.And(`even multiple flag in the system`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					UpdateReleaseFlagRolloutPercentage(t, `b`, 20)
					UpdateReleaseFlagRolloutPercentage(t, `c`, 30)
				})

				s.Then(`flags received back`, func(t *testcase.T) {
					r := subject(t)
					require.Equal(t, 200, r.Code)
					require.Equal(t, "application/json", r.Header().Get(`Content-Type`))

					var resps []map[string]interface{}
					IsJsonResponse(t, r, &resps)
					require.Equal(t, 3, len(resps))
				})
			})
		})
	})

}
