package httpapi_test

import (
	"bytes"
	"encoding/json"
	"github.com/toggler-io/toggler/lib/go/client"
	"github.com/toggler-io/toggler/lib/go/client/operations"
	"github.com/toggler-io/toggler/lib/go/models"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/extintf/httpintf/httpapi"
	. "github.com/toggler-io/toggler/testing"
	"github.com/stretchr/testify/require"
)

func TestServeMux_IsFeatureEnabledFor(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	subject := func(t *testcase.T) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		NewServeMux(t).ServeHTTP(rr, t.I(`request`).(*http.Request))
		return rr
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

				var resp struct {
					Enrollment bool `json:"enrollment"`
				}

				IsJsonResponse(t, r, &resp)
				require.Equal(t, true, resp.Enrollment)
			})
		})

		s.And(`pilot is not`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				SpecPilotEnrolmentIs(t, false)
			})

			s.Then(`the request will be marked as forbidden`, func(t *testcase.T) {
				r := subject(t)

				require.Equal(t, 200, r.Code)

				var resp struct {
					Enrollment bool `json:"enrollment"`
				}

				IsJsonResponse(t, r, &resp)
				require.Equal(t, false, resp.Enrollment)
			})
		})

	}

	s.When(`params sent trough query string content`, func(s *testcase.Spec) {

		s.Let(`request`, func(t *testcase.T) interface{} {
			u, err := url.Parse(`/release/is-feature-enabled.json`)
			require.Nil(t, err)

			q := u.Query()
			q.Set(`feature`, GetReleaseFlagName(t))
			q.Set(`id`, GetExternalPilotID(t))
			u.RawQuery = q.Encode()

			return httptest.NewRequest(http.MethodGet, u.String(), bytes.NewBuffer([]byte{}))
		})

		sharedSpec(s)

	})

	s.When(`params sent trough json body content`, func(s *testcase.Spec) {

		s.Let(`request`, func(t *testcase.T) interface{} {
			u, err := url.Parse(`/release/is-feature-enabled.json`)
			require.Nil(t, err)
			payload := bytes.NewBuffer([]byte{})
			jsonenc := json.NewEncoder(payload)
			require.Nil(t, jsonenc.Encode(httpapi.IsFeatureEnabledRequestPayload{
				Feature: GetReleaseFlagName(t),
				PilotID: GetExternalPilotID(t),
			}))

			r := httptest.NewRequest(http.MethodGet, u.String(), payload)
			r.Header.Set(`Content-Type`, `application/json`)
			return r
		})

		s.Let(`swagger.client.params`, func(t *testcase.T) interface{} {
			p := operations.NewIsFeatureEnabledParams()
			p.Body = &models.IsFeatureEnabledRequestPayload{}
			pilotExtID := GetExternalPilotID(t)
			ffName := GetReleaseFlagName(t)
			p.Body.PilotID = &pilotExtID
			p.Body.Feature = &ffName
			return p
		})

		sharedSpec(s)

	})

	s.Test(`swagger integration`, func(t *testcase.T) {

		enr := rand.Intn(2) == 0
		if enr {
			GetReleaseFlag(t).Rollout.Strategy.Percentage = 100
		}

		require.Nil(t, GetStorage(t).Save(CTX(t), GetReleaseFlag(t)))

		s := httptest.NewServer(http.StripPrefix(`/api/v1`, NewServeMux(t)))
		defer s.Close()

		p := operations.NewIsFeatureEnabledParams()
		p.Body = &models.IsFeatureEnabledRequestPayload{}
		pilotExtID := GetExternalPilotID(t)
		ffName := GetReleaseFlagName(t)
		p.Body.PilotID = &pilotExtID
		p.Body.Feature = &ffName

		tc := client.DefaultTransportConfig()
		u, _ := url.Parse(s.URL)
		tc.Host = u.Host
		tc.Schemes = []string{`http`}

		c := client.NewHTTPClientWithConfig(nil, tc)

		resp, err := c.Operations.IsFeatureEnabled(p)
		if err != nil {
			t.Fatal(err.Error())
		}

		require.NotNil(t, resp)
		require.NotNil(t, resp.Payload)
		require.Equal(t, enr, resp.Payload.Enrollment)

	})
}
