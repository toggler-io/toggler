package httpapi_test

import (
	"bytes"
	"encoding/json"
	"github.com/toggler-io/toggler/extintf/httpintf/httpapi"
	"github.com/toggler-io/toggler/lib/go/client"
	"github.com/toggler-io/toggler/lib/go/client/operations"
	"github.com/toggler-io/toggler/lib/go/models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/adamluzsi/testcase"
	. "github.com/toggler-io/toggler/testing"
	"github.com/stretchr/testify/require"
)

func TestServeMux_ClientConfig(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	subject := func(t *testcase.T) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := t.I(`request`).(*http.Request)
		NewServeMux(t).ServeHTTP(w, r)
		return w
	}

	stateIs := func(t *testcase.T, key string, state bool, states map[string]bool) {
		value, ok := states[key]
		require.True(t, ok)
		require.Equal(t, state, value)
	}

	SetupSpecCommonVariables(s)

	sharedSpec := func(s *testcase.Spec) {

		s.And(`pilot is enrolled`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				SpecPilotEnrolmentIs(t, true)
			})

			s.Then(`the request will be accepted with OK`, func(t *testcase.T) {
				r := subject(t)
				require.Equal(t, 200, r.Code)
				var body httpapi.RolloutClientConfigResponseBody
				IsJsonResponse(t, r, &body)
				stateIs(t, GetFeatureFlagName(t), true, body.States)
				stateIs(t, `yet-unknown-feature`, false, body.States)
			})
		})

		s.And(`pilot is not`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { SpecPilotEnrolmentIs(t, false) })

			s.Then(`the request will include values about toggles being flipped off`, func(t *testcase.T) {
				r := subject(t)
				require.Equal(t, 200, r.Code)
				var body httpapi.RolloutClientConfigResponseBody
				IsJsonResponse(t, r, &body)
				stateIs(t, GetFeatureFlagName(t), false, body.States)
				stateIs(t, `yet-unknown-feature`, false, body.States)
			})
		})

	}

	s.When(`params sent trough`, func(s *testcase.Spec) {
		s.Context(`query string`, func(s *testcase.Spec) {
			s.And(`the feature query string key`, func(s *testcase.Spec) {
				s.Let(`request`, func(t *testcase.T) interface{} {
					u, err := url.Parse(`/rollout/config.json`)
					require.Nil(t, err)

					q := u.Query()
					q.Set(t.I(`feature query string key`).(string), GetFeatureFlagName(t))
					q.Add(t.I(`feature query string key`).(string), `yet-unknown-feature`)
					q.Set(`id`, GetExternalPilotID(t))
					u.RawQuery = q.Encode()

					return httptest.NewRequest(http.MethodGet, u.String(), bytes.NewBuffer([]byte{}))
				})

				s.Context(`is "feature"`, func(s *testcase.Spec) {
					s.Let(`feature query string key`, func(t *testcase.T) interface{} {
						return `feature`
					})

					sharedSpec(s)
				})

				s.Context(`is "feature[]"`, func(s *testcase.Spec) {
					s.Let(`feature query string key`, func(t *testcase.T) interface{} {
						return `feature[]`
					})

					sharedSpec(s)
				})
			})
		})

		s.Context(`payload serialized as json`, func(s *testcase.Spec) {
			s.Let(`request`, func(t *testcase.T) interface{} {
				u, err := url.Parse(`/rollout/config.json`)
				require.Nil(t, err)
				payload := bytes.NewBuffer([]byte{})
				jsonenc := json.NewEncoder(payload)
				require.Nil(t, jsonenc.Encode(httpapi.RolloutClientConfigRequestBody{
					PilotID:  GetExternalPilotID(t),
					Features: []string{GetFeatureFlagName(t), "yet-unknown-feature"},
				}))

				r := httptest.NewRequest(http.MethodGet, u.String(), payload)
				r.Header.Set(`Content-Type`, `application/json`)
				return r
			})

			sharedSpec(s)
		})

	})

	s.Test(`swagger integration`, func(t *testcase.T) {

		require.Nil(t, GetStorage(t).Save(CTX(t), GetFeatureFlag(t)))
		require.Nil(t, GetStorage(t).Save(CTX(t), GetPilot(t)))

		s := httptest.NewServer(http.StripPrefix(`/api/v1`, NewServeMux(t)))
		defer s.Close()

		p := operations.NewRolloutClientConfigParams()
		p.Body = &models.RolloutClientConfigRequestBody{}
		p.Body.PilotID = &GetPilot(t).ExternalID
		p.Body.Features = []string{GetFeatureFlagName(t)}

		tc := client.DefaultTransportConfig()
		u, _ := url.Parse(s.URL)
		tc.Host = u.Host
		tc.Schemes = []string{`http`}

		c := client.NewHTTPClientWithConfig(nil, tc)

		resp, err := c.Operations.RolloutClientConfig(p)
		if err != nil {
			t.Fatal(err.Error())
		}

		require.NotNil(t, resp)
		require.NotNil(t, resp.Payload)
		require.Equal(t, GetPilotEnrollment(t), resp.Payload.States[GetFeatureFlagName(t)])

	})

}
