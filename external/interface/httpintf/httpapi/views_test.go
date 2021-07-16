package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/adamluzsi/testcase/httpspec"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/external/interface/httpintf"
	"github.com/toggler-io/toggler/external/interface/httpintf/httpapi"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client/pilot"
	sh "github.com/toggler-io/toggler/spechelper"
)

func TestViewsController(t *testing.T) {
	s := sh.NewSpec(t)
	s.Parallel()

	HandlerLet(s, func(t *testcase.T) http.Handler { return NewHandler(t) })
	Context.Let(s, func(t *testcase.T) interface{} { return sh.ContextGet(t) })

	s.Describe(`GET /v/config - GetPilotConfig`, SpecViewsControllerClientConfig)
}

func SpecViewsControllerClientConfig(s *testcase.Spec) {
	Method.LetValue(s, `GET`)
	Path.LetValue(s, `/v/config`)

	var onSuccess = func(t *testcase.T) httpapi.GetPilotConfigResponse {
		r := ServeHTTP(t)
		require.Equal(t, 200, r.Code, r.Body.String())
		var resp httpapi.GetPilotConfigResponse
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
				sh.SpecPilotEnrolmentIs(t, true)
			})

			s.Then(`the request will be accepted with OK`, func(t *testcase.T) {
				response := onSuccess(t)
				stateIs(t, sh.ExampleReleaseFlag(t).Name, true, response.Body.Release.Flags)
				stateIs(t, `yet-unknown-feature`, false, response.Body.Release.Flags)
			})
		})

		s.And(`pilot is not`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { sh.SpecPilotEnrolmentIs(t, false) })

			s.Then(`the request will include values about toggles being flipped off`, func(t *testcase.T) {
				response := onSuccess(t)
				stateIs(t, sh.ExampleReleaseFlag(t).Name, false, response.Body.Release.Flags)
				stateIs(t, `yet-unknown-feature`, false, response.Body.Release.Flags)
			})
		})
	}

	s.When(`params sent trough`, func(s *testcase.Spec) {
		s.Context(`query string`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				QueryGet(t).Set(`env`, sh.ExampleDeploymentEnvironment(t).ID)
				QueryGet(t).Set(t.I(`feature query string key`).(string), sh.ExampleReleaseFlag(t).Name)
				QueryGet(t).Add(t.I(`feature query string key`).(string), `yet-unknown-feature`)
				QueryGet(t).Set(`external_id`, sh.ExampleExternalPilotID(t))
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
			Body.Let(s, func(t *testcase.T) interface{} {
				HeaderGet(t).Set(`Content-Type`, `application/json`)
				payload := bytes.NewBuffer([]byte{})
				jsonenc := json.NewEncoder(payload)
				var confReq httpapi.GetPilotConfigRequest
				confReq.Body.PilotExtID = sh.ExampleExternalPilotID(t)
				confReq.Body.DeploymentEnvironmentAlias = sh.ExampleDeploymentEnvironment(t).ID
				confReq.Body.ReleaseFlags = []string{sh.ExampleReleaseFlag(t).Name, "yet-unknown-feature"}
				require.Nil(t, jsonenc.Encode(confReq.Body))
				return payload
			})

			sharedSpec(s)
		})
	})

	s.Context(`E2E`, func(s *testcase.Spec) {
		s.Tag(sh.TagBlackBox)

		s.Test(`swagger`, func(t *testcase.T) {
			sm, err := httpintf.NewServeMux(sh.ExampleUseCases(t))
			require.Nil(t, err)

			s := httptest.NewServer(sm)
			defer s.Close()

			p := pilot.NewGetPilotConfigParams()
			p.Body.PilotExtID = &sh.ExampleReleaseManualPilotEnrollment(t).PublicID
			p.Body.ReleaseFlags = []string{sh.ExampleReleaseFlag(t).Name}
			p.Body.DeploymentEnvironmentAlias = &sh.ExampleDeploymentEnvironment(t).ID

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			resp, err := c.Pilot.GetPilotConfig(p, publicAuth(t))
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Payload)
			require.Equal(t, sh.ExampleReleaseManualPilotEnrollment(t).IsParticipating, resp.Payload.Release.Flags[sh.ExampleReleaseFlag(t).Name])
		})
	})
}
