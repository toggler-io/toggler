package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/adamluzsi/testcase/httpspec"

	"github.com/toggler-io/toggler/extintf/httpintf"
	"github.com/toggler-io/toggler/extintf/httpintf/httpapi"
	"github.com/toggler-io/toggler/lib/go/client"
	"github.com/toggler-io/toggler/lib/go/client/release_flag"
	"github.com/toggler-io/toggler/usecases"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	. "github.com/toggler-io/toggler/testing"
)

func TestViewsController(t *testing.T) {
	Debug = true
	s := testcase.NewSpec(t)
	s.Parallel()

	SetupSpecCommonVariables(s)
	GivenThisIsAJSONAPI(s)

	LetHandler(s, func(t *testcase.T) http.Handler { return NewHandler(t) })

	s.Describe(`GET /v/client-config - ClientConfig`, SpecViewsControllerClientConfig)

	s.Test(`swagger integration`, func(t *testcase.T) {
		require.Nil(t, GetStorage(t).Create(CTX(t), GetReleaseFlag(t)))
		require.Nil(t, GetStorage(t).Create(CTX(t), GetPilot(t)))

		sm, err := httpintf.NewServeMux(usecases.NewUseCases(GetStorage(t)))
		require.Nil(t, err)

		s := httptest.NewServer(sm)
		defer s.Close()

		p := release_flag.NewGetClientConfigParams()
		p.Body.PilotExtID = &GetPilot(t).ExternalID
		p.Body.ReleaseFlags = []string{GetReleaseFlagName(t)}

		tc := client.DefaultTransportConfig()
		u, _ := url.Parse(s.URL)
		tc.Host = u.Host
		tc.Schemes = []string{`http`}

		c := client.NewHTTPClientWithConfig(nil, tc)

		resp, err := c.ReleaseFlag.GetClientConfig(p)
		if err != nil {
			t.Fatal(err.Error())
		}

		require.NotNil(t, resp)
		require.NotNil(t, resp.Payload)
		require.Equal(t, GetPilotEnrollment(t), resp.Payload.Release.Flags[GetReleaseFlagName(t)])
	})
}

func SpecViewsControllerClientConfig(s *testcase.Spec) {
	LetMethodValue(s, `GET`)
	LetPathValue(s, `/v/client-config`)

	var onSuccess = func(t *testcase.T) httpapi.ClientConfigResponse {
		r := ServeHTTP(t)
		require.Equal(t, 200, r.Code)
		var resp httpapi.ClientConfigResponse
		IsJsonResponse(t, r, &resp.Body)
		return resp
	}

	var stateIs = func(t *testcase.T, key string, state bool, states map[string]bool) {
		value, ok := states[key]
		require.True(t, ok)
		require.Equal(t, state, value)
	}

	sharedSpec := func(s *testcase.Spec) {
		s.And(`pilot is enrolled`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				SpecPilotEnrolmentIs(t, true)
			})

			s.Then(`the request will be accepted with OK`, func(t *testcase.T) {
				response := onSuccess(t)
				stateIs(t, GetReleaseFlagName(t), true, response.Body.Release.Flags)
				stateIs(t, `yet-unknown-feature`, false, response.Body.Release.Flags)
			})
		})

		s.And(`pilot is not`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { SpecPilotEnrolmentIs(t, false) })

			s.Then(`the request will include values about toggles being flipped off`, func(t *testcase.T) {
				response := onSuccess(t)
				stateIs(t, GetReleaseFlagName(t), false, response.Body.Release.Flags)
				stateIs(t, `yet-unknown-feature`, false, response.Body.Release.Flags)
			})
		})
	}

	s.When(`params sent trough`, func(s *testcase.Spec) {
		s.Context(`query string`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				Query(t).Set(t.I(`feature query string key`).(string), GetReleaseFlagName(t))
				Query(t).Add(t.I(`feature query string key`).(string), `yet-unknown-feature`)
				Query(t).Set(`id`, GetExternalPilotID(t))
			})

			s.Context(`is "release_flags"`, func(s *testcase.Spec) {
				s.Let(`feature query string key`, func(t *testcase.T) interface{} { return `release_flags` })

				sharedSpec(s)
			})

			s.Context(`is "release_flags[]"`, func(s *testcase.Spec) {
				s.Let(`feature query string key`, func(t *testcase.T) interface{} { return `release_flags[]` })

				sharedSpec(s)
			})
		})

		s.Context(`payload serialized as json`, func(s *testcase.Spec) {
			LetBody(s, func(t *testcase.T) interface{} {
				Header(t).Set(`Content-Type`, `application/json`)
				payload := bytes.NewBuffer([]byte{})
				jsonenc := json.NewEncoder(payload)
				var confReq httpapi.ClientConfigRequest
				confReq.Body.PilotExtID = GetExternalPilotID(t)
				confReq.Body.ReleaseFlags = []string{GetReleaseFlagName(t), "yet-unknown-feature"}
				require.Nil(t, jsonenc.Encode(confReq.Body))
				return payload
			})

			sharedSpec(s)
		})
	})
}
