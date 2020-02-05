package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/toggler-io/toggler/lib/go/client"
	"github.com/toggler-io/toggler/lib/go/client/release_flag"

	"github.com/toggler-io/toggler/extintf/httpintf/httpapi"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	. "github.com/toggler-io/toggler/testing"
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
				var body httpapi.ClientConfigResponseBody
				IsJsonResponse(t, r, &body)
				stateIs(t, GetReleaseFlagName(t), true, body.Release.Flags)
				stateIs(t, `yet-unknown-feature`, false, body.Release.Flags)
			})
		})

		s.And(`pilot is not`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { SpecPilotEnrolmentIs(t, false) })

			s.Then(`the request will include values about toggles being flipped off`, func(t *testcase.T) {
				r := subject(t)
				require.Equal(t, 200, r.Code)
				var body httpapi.ClientConfigResponseBody
				IsJsonResponse(t, r, &body)
				stateIs(t, GetReleaseFlagName(t), false, body.Release.Flags)
				stateIs(t, `yet-unknown-feature`, false, body.Release.Flags)
			})
		})

	}

	s.When(`params sent trough`, func(s *testcase.Spec) {
		s.Context(`query string`, func(s *testcase.Spec) {
			s.And(`the feature query string key`, func(s *testcase.Spec) {
				s.Let(`request`, func(t *testcase.T) interface{} {
					u, err := url.Parse(`/client/config.json`)
					require.Nil(t, err)

					q := u.Query()
					q.Set(t.I(`feature query string key`).(string), GetReleaseFlagName(t))
					q.Add(t.I(`feature query string key`).(string), `yet-unknown-feature`)
					q.Set(`id`, GetExternalPilotID(t))
					u.RawQuery = q.Encode()

					return httptest.NewRequest(http.MethodGet, u.String(), bytes.NewBuffer([]byte{}))
				})

				s.Context(`is "release_flags"`, func(s *testcase.Spec) {
					s.Let(`feature query string key`, func(t *testcase.T) interface{} {
						return `release_flags`
					})

					sharedSpec(s)
				})

				s.Context(`is "release_flags[]"`, func(s *testcase.Spec) {
					s.Let(`feature query string key`, func(t *testcase.T) interface{} {
						return `release_flags[]`
					})

					sharedSpec(s)
				})
			})
		})

		s.Context(`payload serialized as json`, func(s *testcase.Spec) {
			s.Let(`request`, func(t *testcase.T) interface{} {
				u, err := url.Parse(`/client/config.json`)
				require.Nil(t, err)
				payload := bytes.NewBuffer([]byte{})
				jsonenc := json.NewEncoder(payload)

				var confReq httpapi.ClientConfigRequest
				confReq.Body.PilotExtID = GetExternalPilotID(t)
				confReq.Body.ReleaseFlags = []string{GetReleaseFlagName(t), "yet-unknown-feature"}
				require.Nil(t, jsonenc.Encode(confReq.Body))

				r := httptest.NewRequest(http.MethodGet, u.String(), payload)
				r.Header.Set(`Content-Type`, `application/json`)
				return r
			})

			sharedSpec(s)
		})

	})

	s.Test(`swagger integration`, func(t *testcase.T) {

		require.Nil(t, GetStorage(t).Create(CTX(t), GetReleaseFlag(t)))
		require.Nil(t, GetStorage(t).Create(CTX(t), GetPilot(t)))

		s := httptest.NewServer(http.StripPrefix(`/api/v1`, NewServeMux(t)))
		defer s.Close()

		p := release_flag.NewClientConfigParams()
		p.Body.PilotExtID = &GetPilot(t).ExternalID
		p.Body.ReleaseFlags = []string{GetReleaseFlagName(t)}

		tc := client.DefaultTransportConfig()
		u, _ := url.Parse(s.URL)
		tc.Host = u.Host
		tc.Schemes = []string{`http`}

		c := client.NewHTTPClientWithConfig(nil, tc)

		resp, err := c.ReleaseFlag.ClientConfig(p)
		if err != nil {
			t.Fatal(err.Error())
		}

		require.NotNil(t, resp)
		require.NotNil(t, resp.Payload)
		require.Equal(t, GetPilotEnrollment(t), resp.Payload.Release.Flags[GetReleaseFlagName(t)])

	})

}
