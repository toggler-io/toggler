package httpapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/testcase"
	hs "github.com/adamluzsi/testcase/httpspec"
	"github.com/stretchr/testify/require"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/external/interface/httpintf/httpapi"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client"
	swagger "github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client/pilot"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/models"
	sh "github.com/toggler-io/toggler/spechelper"
)

func TestReleasePilotController(t *testing.T) {
	s := sh.NewSpec(t)
	s.Parallel()

	hs.Context.Let(s, func(t *testcase.T) interface{} { return sh.ContextGet(t) })
	hs.HandlerLet(s, func(t *testcase.T) http.Handler { return httpapi.NewReleasePilotHandler(sh.ExampleUseCases(t)) })
	hs.ContentTypeIsJSON(s)
	sh.GivenHTTPRequestHasAppToken(s)

	s.Describe("POST / - create", SpecReleasePilotControllerCreate)
	s.Describe("GET / - list", SpecReleasePilotControllerList)
	s.Describe("PUT /:pilot_id - update", SpecReleasePilotControllerUpdate)
	s.Describe("DELETE /:pilot_id - delete", SpecReleasePilotControllerDelete)
}

func SpecReleasePilotControllerCreate(s *testcase.Spec) {
	hs.Method.LetValue(s, http.MethodPost)
	hs.Path.LetValue(s, `/`)

	var onSuccess = func(t *testcase.T) (resp httpapi.CreateReleasePilotResponse) {
		rr := hs.ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.After(func(t *testcase.T) {
		require.Nil(t, sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).DeleteAll(sh.ContextGet(t)))
	})

	pilot := s.Let(`release-pilot`, func(t *testcase.T) interface{} {
		rf := sh.NewFixtureFactory(t).Fixture(release.Pilot{}, sh.ContextGet(t)).(release.Pilot)
		return &rf
	})
	pilotGet := func(t *testcase.T) *release.Pilot {
		return pilot.Get(t).(*release.Pilot)
	}

	hs.Body.Let(s, func(t *testcase.T) interface{} {
		var req httpapi.CreateReleasePilotRequest
		req.Body.Pilot = *pilotGet(t)
		return req.Body
	})

	s.Then(`call succeed`, func(t *testcase.T) {
		require.Equal(t, 200, hs.ServeHTTP(t).Code)
	})

	s.Then(`pilot stored in the system`, func(t *testcase.T) {
		onSuccess(t)
		v := pilotGet(t)

		var pilots []release.Pilot
		pilotsIter := sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).FindAll(sh.ContextGet(t))
		require.Nil(t, iterators.Collect(pilotsIter, &pilots))
		require.Len(t, pilots, 1)

		actual := pilots[0]
		require.NotEmpty(t, actual.ID)
		actual.ID = ""
		require.Equal(t, *v, actual)
	})

	s.Then(`it returns pilot in the response`, func(t *testcase.T) {
		resp := onSuccess(t)
		require.NotEmpty(t, resp.Body.Pilot.ID)
		resp.Body.Pilot.ID = ""
		require.Equal(t, *pilotGet(t), resp.Body.Pilot)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`for example, reference ids are empty`)
			pilotGet(t).FlagID = ""
			pilotGet(t).EnvironmentID = ""
		})

		s.Then(`it will return with failure`, func(t *testcase.T) {
			rr := hs.ServeHTTP(t)
			require.Equal(t, http.StatusBadRequest, rr.Code)

			var resp httpapi.ErrorResponse
			require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body), rr.Body.String())
			require.NotEmpty(t, resp.Body.Error.Message)
		})
	})

	s.Test(`confirm swagger documentation`, func(t *testcase.T) {
		s := NewHTTPServer(t)

		// TODO: ensure validation
		p := swagger.NewCreateReleasePilotParams()
		p.Body.Pilot = &models.Pilot{
			EnvironmentID:   sh.ExampleDeploymentEnvironment(t).ID,
			FlagID:          sh.ExampleReleaseFlag(t).ID,
			IsParticipating: t.Random.Bool(),
			PublicID:        t.Random.String(),
		}

		tc := client.DefaultTransportConfig()
		u, _ := url.Parse(s.URL)
		tc.Host = u.Host
		tc.Schemes = []string{`http`}

		c := client.NewHTTPClientWithConfig(nil, tc)

		resp, err := c.Pilot.CreateReleasePilot(p, protectedAuth(t))
		if err != nil {
			t.Fatal(err.Error())
		}

		require.NotNil(t, resp)
		require.NotNil(t, resp.Payload)
	})
}

func SpecReleasePilotControllerList(s *testcase.Spec) {
	hs.Method.LetValue(s, http.MethodGet)
	hs.Path.LetValue(s, `/`)

	var onSuccess = func(t *testcase.T) httpapi.ListReleasePilotResponse {
		var resp httpapi.ListReleasePilotResponse
		rr := hs.ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.And(`no pilot present in the system`, func(s *testcase.Spec) {
		sh.NoReleasePilotPresentInTheStorage(s)

		s.Then(`empty result received`, func(t *testcase.T) {
			require.Empty(t, onSuccess(t).Body.Pilots)
		})
	})

	s.And(`release pilot is present in the system`, func(s *testcase.Spec) {
		pilot1 := sh.GivenWeHaveReleasePilot(s, `pilot-1`)

		s.Then(`pilot received back`, func(t *testcase.T) {
			resp := onSuccess(t)
			require.Len(t, resp.Body.Pilots, 1)
			require.Contains(t, resp.Body.Pilots, *sh.ReleasePilotGet(t, pilot1))
		})

		s.And(`even multiple pilot in the system`, func(s *testcase.Spec) {
			pilot2 := sh.GivenWeHaveReleasePilot(s, `pilot-2`)

			s.Then(`the pilots will be received back`, func(t *testcase.T) {
				resp := onSuccess(t)
				require.Len(t, resp.Body.Pilots, 2)
				require.Contains(t, resp.Body.Pilots, *sh.ReleasePilotGet(t, pilot2))
			})
		})
	})

	// TODO: swagger integration test
}

func SpecReleasePilotControllerUpdate(s *testcase.Spec) {
	sh.GivenHTTPRequestHasAppToken(s)
	hs.Method.LetValue(s, http.MethodPut)

	pilot := sh.GivenWeHaveReleasePilot(s, `release-pilot`)

	hs.Path.Let(s, func(t *testcase.T) interface{} {
		return fmt.Sprintf(`/%s`, sh.ReleasePilotGet(t, pilot).ID)
	})

	updatedPilot := s.Let(`updated-release-pilot`, func(t *testcase.T) interface{} {
		rf := sh.NewFixtureFactory(t).Fixture(release.Pilot{}, sh.ContextGet(t)).(release.Pilot)
		rf.ID = sh.ReleasePilotGet(t, pilot).ID
		return &rf
	})

	hs.Body.Let(s, func(t *testcase.T) interface{} {
		var req httpapi.CreateReleasePilotRequest
		req.Body.Pilot = *sh.ReleasePilotGet(t, updatedPilot)
		return req.Body
	})

	var onSuccess = func(t *testcase.T) httpapi.UpdateReleasePilotResponse {
		rr := hs.ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code)
		var resp httpapi.UpdateReleasePilotResponse
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.Then(`pilot is updated in the system`, func(t *testcase.T) {
		onSuccess(t)

		var pilots []release.Pilot
		pilotsIter := sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).FindAll(sh.ContextGet(t))
		require.Nil(t, iterators.Collect(pilotsIter, &pilots))
		require.Len(t, pilots, 1)

		actual := pilots[0]
		require.NotEmpty(t, actual.ID)
		require.Equal(t, *sh.ReleasePilotGet(t, updatedPilot), actual)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			v := sh.ReleasePilotGet(t, updatedPilot)
			v.FlagID = ""
			v.EnvironmentID = ""
			updatedPilot.Set(t, v)
		})

		s.Then(`it will return with failure`, func(t *testcase.T) {
			rr := hs.ServeHTTP(t)
			require.Equal(t, http.StatusBadRequest, rr.Code)

			var resp httpapi.ErrorResponse
			require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body), rr.Body.String())
			require.NotEmpty(t, resp.Body.Error.Message)
		})
	})

	s.Context(`E2E`, func(s *testcase.Spec) {
		s.Tag(sh.TagBlackBox)

		s.Test(`swagger`, func(t *testcase.T) {
			s := NewHTTPServer(t)

			id := sh.ReleasePilotGet(t, pilot).ID
			contracts.IsFindable(t, release.Pilot{}, sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)), sh.ContextGet(t), id)

			p := swagger.NewUpdateReleasePilotParams()
			p.PilotID = id
			p.Body.Pilot = &models.Pilot{
				ID:              id,
				EnvironmentID:   sh.ExampleDeploymentEnvironment(t).ID,
				FlagID:          sh.ExampleReleaseFlag(t).ID,
				IsParticipating: t.Random.Bool(),
				PublicID:        t.Random.String(),
			}

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			resp, err := c.Pilot.UpdateReleasePilot(p, protectedAuth(t))
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Payload)
		})
	})
}

func SpecReleasePilotControllerDelete(s *testcase.Spec) {
	sh.GivenHTTPRequestHasAppToken(s)
	hs.Method.LetValue(s, http.MethodDelete)

	pilot := sh.GivenWeHaveReleasePilot(s, "release-pilot")

	hs.Path.Let(s, func(t *testcase.T) interface{} {
		return fmt.Sprintf(`/%s`, sh.ReleasePilotGet(t, pilot).ID)
	})

	var onSuccess = func(t *testcase.T) {
		rr := hs.ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code)
	}

	s.Then(`flag is deleted from the system`, func(t *testcase.T) {
		onSuccess(t)
		deletedPilot := sh.ReleasePilotGet(t, pilot)
		found, err := sh.StorageGet(t).ReleasePilot(sh.ContextGet(t)).FindByID(sh.ContextGet(t), &release.Pilot{}, deletedPilot.ID)
		require.Nil(t, err)
		require.False(t, found)
	})

	s.Context(`E2E`, func(s *testcase.Spec) {
		s.Tag(sh.TagBlackBox)

		s.Test(`swagger`, func(t *testcase.T) {
			s := NewHTTPServer(t)

			p := swagger.NewDeleteReleasePilotParams()
			p.PilotID = sh.ReleasePilotGet(t, pilot).ID

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			_, err := c.Pilot.DeleteReleasePilot(p, protectedAuth(t))
			require.Nil(t, err)
		})
	})
}
