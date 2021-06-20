package httpapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/testcase/httpspec"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/external/interface/httpintf"
	"github.com/toggler-io/toggler/external/interface/httpintf/httpapi"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client"
	swagger "github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client/flag"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/models"
	sh "github.com/toggler-io/toggler/spechelper"
)

func TestReleaseFlagController(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()
	sh.SetUp(s)

	HandlerLet(s, func(t *testcase.T) http.Handler {
		return httpapi.NewReleaseFlagHandler(sh.ExampleUseCases(t))
	})

	ContentTypeIsJSON(s)

	Context.Let(s, func(t *testcase.T) interface{} {
		return sh.ContextGet(t)
	})

	s.Describe(`POST / - create release flag`, SpecReleaseFlagControllerCreate)
	s.Describe(`GET / - list release flags`, SpecReleaseFlagControllerList)

	s.Context(`given we have a release flag in the system`, func(s *testcase.Spec) {
		sh.GivenWeHaveReleaseFlag(s, `release-flag`)

		var andFlagIdentifierProvided = func(s *testcase.Spec, context func(s *testcase.Spec)) {
			s.And(`release flag identifier provided as the external ID`, func(s *testcase.Spec) {
				s.Let(`id`, func(t *testcase.T) interface{} {
					return sh.GetReleaseFlag(t, `release-flag`).ID
				})

				context(s)
			})

			//s.And(`release flag identifier provided as url normalized release flag name`, func(s *testcase.Spec) {
			//	// TODO add implementation to "alias id that can be guessed from the flag name"
			//
			//	s.Let(`id`, func(t *testcase.T) interface{} {
			//		return GetReleaseFlag(t, `release-flag`).Name
			//	})
			//
			//	context(s)
			//})
		}

		andFlagIdentifierProvided(s, func(s *testcase.Spec) {
			s.Describe(`PUT|PATCH /{id} - update a release flag`,
				SpecReleaseFlagControllerUpdate)
		})
	})
}

func SpecReleaseFlagControllerCreate(s *testcase.Spec) {
	Method.LetValue(s, http.MethodPost)
	Path.LetValue(s, `/`)
	sh.GivenHTTPRequestHasAppToken(s)

	var onSuccess = func(t *testcase.T) (resp httpapi.CreateReleaseFlagResponse) {
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.After(func(t *testcase.T) {
		require.Nil(t, sh.StorageGet(t).ReleaseFlag(sh.ContextGet(t)).DeleteAll(sh.ContextGet(t)))
	})

	s.Let(`release-flag`, func(t *testcase.T) interface{} {
		return sh.FixtureFactory{}.Create(release.Flag{}).(*release.Flag)
	})

	Body.Let(s, func(t *testcase.T) interface{} {
		var req httpapi.CreateReleaseFlagRequest
		req.Body.Flag = *t.I(`release-flag`).(*release.Flag)
		return req.Body
	})

	s.Then(`call succeed`, func(t *testcase.T) {
		require.Equal(t, 200, ServeHTTP(t).Code)
	})

	s.Then(`flag stored in the system`, func(t *testcase.T) {
		onSuccess(t)
		rfv := t.I(`release-flag`).(*release.Flag)
		actualReleaseFlag := sh.FindStoredReleaseFlagByName(t, rfv.Name)
		require.Equal(t, rfv.Name, actualReleaseFlag.Name)
	})

	s.Then(`it returns flag in the response`, func(t *testcase.T) {
		resp := onSuccess(t)
		flag := *sh.FindStoredReleaseFlagByName(t, t.I(`release-flag`).(*release.Flag).Name)
		require.Equal(t, flag.Name, resp.Body.Flag.Name)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`for example name is empty`)
			t.I(`release-flag`).(*release.Flag).Name = ``
		})

		s.Then(`it will return with failure`, func(t *testcase.T) {
			rr := ServeHTTP(t)
			require.Equal(t, http.StatusBadRequest, rr.Code)

			var resp httpapi.ErrorResponse
			require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body), rr.Body.String())
			require.NotEmpty(t, resp.Body.Error.Message)
		})
	})

	s.Context(`E2E`, func(s *testcase.Spec) {
		s.Tag(sh.TagBlackBox)

		s.Test(`swagger`, func(t *testcase.T) {
			sm, err := httpintf.NewServeMux(sh.ExampleUseCases(t))
			require.Nil(t, err)

			s := httptest.NewServer(sm)
			defer s.Close()

			// TODO: ensure validation
			p := swagger.NewCreateReleaseFlagParams()
			p.Body.Flag = &models.Flag{
				Name: fixtures.Random.String(),
			}

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			resp, err := c.Flag.CreateReleaseFlag(p, protectedAuth(t))
			if err != nil {
				t.Fatal(err.Error())
			}

			require.NotNil(t, resp)
			require.NotNil(t, resp.Payload)
		})
	})
}

func SpecReleaseFlagControllerList(s *testcase.Spec) {
	Method.LetValue(s, http.MethodGet)
	Path.LetValue(s, `/`)
	sh.GivenHTTPRequestHasAppToken(s)

	var onSuccess = func(t *testcase.T) httpapi.ListReleaseFlagResponse {
		var resp httpapi.ListReleaseFlagResponse
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.And(`no flag present in the system`, func(s *testcase.Spec) {
		sh.NoReleaseFlagPresentInTheStorage(s)

		s.Then(`empty result received`, func(t *testcase.T) {
			require.Empty(t, onSuccess(t).Body.Flags)
		})
	})

	s.And(`release flag is present in the system`, func(s *testcase.Spec) {
		sh.GivenWeHaveReleaseFlag(s, `feature-1`)
		s.Before(func(t *testcase.T) { sh.GetReleaseFlag(t, `feature-1`) }) // eager load

		s.Then(`flag received back`, func(t *testcase.T) {
			resp := onSuccess(t)
			require.Len(t, resp.Body.Flags, 1)
			require.Contains(t, resp.Body.Flags, *sh.GetReleaseFlag(t, `feature-1`))
		})

		s.And(`even multiple flag in the system`, func(s *testcase.Spec) {
			sh.GivenWeHaveReleaseFlag(s, `feature-2`)
			s.Before(func(t *testcase.T) { sh.GetReleaseFlag(t, `feature-2`) }) // eager load

			s.Then(`the flags will be received back`, func(t *testcase.T) {
				resp := onSuccess(t)

				require.Len(t, resp.Body.Flags, 2)
				require.Contains(t, resp.Body.Flags, *sh.GetReleaseFlag(t, `feature-2`))
			})
		})
	})
}

func SpecReleaseFlagControllerUpdate(s *testcase.Spec) {
	sh.GivenHTTPRequestHasAppToken(s)
	Method.LetValue(s, http.MethodPut)
	Path.Let(s, func(t *testcase.T) interface{} {
		return fmt.Sprintf(`/%s`, t.I(`id`))
	})

	s.Let(`updated-release-flag`, func(t *testcase.T) interface{} {
		rf := sh.FixtureFactory{}.Create(release.Flag{}).(*release.Flag)
		rf.ID = sh.GetReleaseFlag(t, `release-flag`).ID
		return rf
	})

	Body.Let(s, func(t *testcase.T) interface{} {
		var req httpapi.CreateReleaseFlagRequest
		req.Body.Flag = *t.I(`updated-release-flag`).(*release.Flag)
		return req.Body
	})

	var onSuccess = func(t *testcase.T) httpapi.UpdateReleaseFlagResponse {
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code)
		var resp httpapi.UpdateReleaseFlagResponse
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.Then(`flag is updated in the system`, func(t *testcase.T) {
		resp := onSuccess(t)

		updatedReleaseFlagView := t.I(`updated-release-flag`).(*release.Flag)
		stored := sh.FindStoredReleaseFlagByName(t, updatedReleaseFlagView.Name)
		require.Equal(t, resp.Body.Flag, *stored)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`for example the name is empty`)
			t.I(`updated-release-flag`).(*release.Flag).Name = ``
		})

		s.Then(`it will return with failure`, func(t *testcase.T) {
			rr := ServeHTTP(t)
			require.Equal(t, http.StatusBadRequest, rr.Code)

			var resp httpapi.ErrorResponse
			require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body), rr.Body.String())
			require.NotEmpty(t, resp.Body.Error.Message)
		})
	})

	s.Context(`E2E`, func(s *testcase.Spec) {
		s.Tag(sh.TagBlackBox)

		s.Test(`swagger`, func(t *testcase.T) {
			sm, err := httpintf.NewServeMux(sh.ExampleUseCases(t))
			require.Nil(t, err)

			s := httptest.NewServer(sm)
			defer s.Close()

			id := sh.GetReleaseFlag(t, `release-flag`).ID

			// TODO: ensure validation
			p := swagger.NewUpdateReleaseFlagParams()
			p.FlagID = id
			p.Body.Flag = &models.Flag{
				ID:   id,
				Name: fixtures.Random.String(),
			}

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			resp, err := c.Flag.UpdateReleaseFlag(p, protectedAuth(t))
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Payload)
		})
	})
}
