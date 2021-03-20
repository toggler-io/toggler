package httpapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/testcase/httpspec"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/external/interface/httpintf"
	"github.com/toggler-io/toggler/external/interface/httpintf/httpapi"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client"
	swagger "github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client/release"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/models"
	sh "github.com/toggler-io/toggler/spechelper"
)

func TestReleaseRolloutController(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()
	sh.SetUp(s)

	HandlerLet(s, func(t *testcase.T) http.Handler {
		return httpapi.NewReleaseRolloutHandler(sh.ExampleUseCases(t))
	})

	ContentTypeIsJSON(s)

	Context.Let(s, func(t *testcase.T) interface{} {
		ctx := sh.GetContext(t)
		ctx = context.WithValue(ctx, httpapi.DeploymentEnvironmentContextKey{}, *sh.ExampleDeploymentEnvironment(t))
		ctx = context.WithValue(ctx, httpapi.ReleaseFlagContextKey{}, *sh.ExampleReleaseFlag(t))
		return ctx
	})

	s.Describe(`POST / - create release rollout`, SpecReleaseRolloutControllerCreate)
	s.Describe(`GET / - list release rollout`, SpecReleaseRolloutControllerList)
	s.Describe(`DELETE / - delete release rollout`, SpecReleaseRolloutControllerDelete)

	s.Context(`given we have a release rollout in the system`, func(s *testcase.Spec) {
		sh.GivenWeHaveReleaseRollout(s, `rollout`, sh.LetVarExampleReleaseFlag, sh.LetVarExampleDeploymentEnvironment)

		s.And(`release rollout identifier provided as the external ID`, func(s *testcase.Spec) {
			s.Let(`id`, func(t *testcase.T) interface{} {
				return sh.GetReleaseRollout(t, `rollout`).ID
			})

			s.Describe(`PUT|PATCH /{id} - update a release rollout`,
				SpecReleaseRolloutControllerUpdate)
		})
	})
}

func SpecReleaseRolloutControllerDelete(s *testcase.Spec) {
	Method.LetValue(s, http.MethodDelete)
	Path.LetValue(s, `/`)
	sh.GivenHTTPRequestHasAppToken(s)

	s.Test(`TODO: add the ability to delete rollout`, func(t *testcase.T) { t.Skip() })
}

func SpecReleaseRolloutControllerCreate(s *testcase.Spec) {
	Method.LetValue(s, http.MethodPost)
	Path.LetValue(s, `/`)
	sh.GivenHTTPRequestHasAppToken(s)

	var onSuccess = func(t *testcase.T) (resp httpapi.CreateReleaseRolloutResponse) {
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.After(func(t *testcase.T) {
		require.Nil(t, sh.StorageGet(t).DeleteAll(sh.GetContext(t), release.Rollout{}))
		require.Nil(t, sh.StorageGet(t).DeleteAll(sh.GetContext(t), release.Flag{}))
	})

	s.Let(`rollout`, func(t *testcase.T) interface{} {
		r := sh.FixtureFactory{}.Create(release.Rollout{}).(*release.Rollout)
		r.FlagID = sh.ExampleReleaseFlag(t).ID
		r.DeploymentEnvironmentID = sh.ExampleDeploymentEnvironment(t).ID
		r.Plan = *sh.FixtureFactory{}.Create(release.RolloutDecisionByPercentage{}).(*release.RolloutDecisionByPercentage)
		return r
	})

	Body.Let(s, func(t *testcase.T) interface{} {
		r := sh.GetReleaseRollout(t, `rollout`)
		var req httpapi.CreateReleaseRolloutRequest
		t.Log(r)
		req.Body.Rollout.Plan = release.RolloutDefinitionView{Definition: r.Plan}
		return req.Body
	})

	s.Then(`call succeed`, func(t *testcase.T) {
		require.Equal(t, 200, ServeHTTP(t).Code)
	})

	s.Then(`rollout stored in the system`, func(t *testcase.T) {
		onSuccess(t)
		rfv := *sh.GetReleaseRollout(t, `rollout`)
		actualReleaseRollout := FindStoredReleaseRollout(t)
		actualReleaseRollout.ID = ``
		require.Equal(t, rfv, actualReleaseRollout)
	})

	s.Then(`it returns rollout in the response`, func(t *testcase.T) {
		resp := onSuccess(t)
		stored := FindStoredReleaseRollout(t)
		planMap := getPlanMap(t, resp.Body.Rollout)
		expectedPlanMap, err := release.RolloutDefinitionView{}.MarshalMapping(stored.Plan)
		require.Nil(t, err)
		require.Equal(t, expectedPlanMap, planMap)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			p := release.NewRolloutDecisionByPercentage()
			p.Percentage = 120
			sh.GetReleaseRollout(t, `rollout`).Plan = p
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

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			p := swagger.NewCreateReleaseRolloutParams()
			p.FlagID = sh.ExampleReleaseFlag(t).ID
			p.EnvironmentID = sh.ExampleDeploymentEnvironment(t).ID
			p.Body.Rollout = &models.Rollout{
				Plan: release.RolloutDefinitionView{Definition: release.NewRolloutDecisionByPercentage()},
			}

			resp, err := c.Release.CreateReleaseRollout(p, protectedAuth(t))
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Payload)
		})
	})
}

func SpecReleaseRolloutControllerList(s *testcase.Spec) {
	Method.LetValue(s, http.MethodGet)
	Path.LetValue(s, `/`)
	sh.GivenHTTPRequestHasAppToken(s)

	var onSuccess = func(t *testcase.T) httpapi.ListReleaseRolloutResponse {
		var resp httpapi.ListReleaseRolloutResponse
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.And(`no rollout present in the system`, func(s *testcase.Spec) {
		sh.NoReleaseRolloutPresentInTheStorage(s)

		s.Then(`empty result received`, func(t *testcase.T) {
			require.Empty(t, onSuccess(t).Body.Rollouts)
		})
	})

	s.And(`release rollout is present in the system`, func(s *testcase.Spec) {
		sh.GivenWeHaveReleaseRollout(s, `rollout-1`, sh.LetVarExampleReleaseFlag, sh.LetVarExampleDeploymentEnvironment)
		s.Before(func(t *testcase.T) { sh.GetReleaseRollout(t, `rollout-1`) }) // eager load

		s.Then(`rollout received back`, func(t *testcase.T) {
			resp := onSuccess(t)
			require.Len(t, resp.Body.Rollouts, 1)
			rollout := sh.GetReleaseRollout(t, `rollout-1`)
			plan, err := release.RolloutDefinitionView{}.MarshalMapping(rollout.Plan)
			require.Nil(t, err)

			ar := resp.Body.Rollouts[0]
			require.Equal(t, rollout.ID, ar.ID)
			require.Equal(t, rollout.FlagID, ar.FlagID)
			require.Equal(t, rollout.DeploymentEnvironmentID, ar.EnvironmentID)
			require.Equal(t, plan, getPlanMap(t, ar))
		})

		s.And(`even multiple rollout in the system`, func(s *testcase.Spec) {
			sh.GivenWeHaveReleaseRollout(s, `feature-2`, sh.LetVarExampleReleaseFlag, sh.LetVarExampleDeploymentEnvironment)
			s.Before(func(t *testcase.T) { sh.GetReleaseRollout(t, `feature-2`) }) // eager load

			s.Then(`the rollouts will be received back`, func(t *testcase.T) {
				resp := onSuccess(t)
				rollout := sh.GetReleaseRollout(t, `feature-2`)
				require.Len(t, resp.Body.Rollouts, 2)
				plan, err := release.RolloutDefinitionView{}.MarshalMapping(rollout.Plan)
				require.Nil(t, err)

				var ar httpapi.Rollout
				for _, r := range resp.Body.Rollouts {
					if r.ID == rollout.ID {
						ar = r
						break
					}
				}

				require.Equal(t, rollout.ID, ar.ID)
				require.Equal(t, rollout.FlagID, ar.FlagID)
				require.Equal(t, rollout.DeploymentEnvironmentID, ar.EnvironmentID)
				require.Equal(t, plan, getPlanMap(t, ar))
			})
		})
	})
}

func SpecReleaseRolloutControllerUpdate(s *testcase.Spec) {
	sh.GivenHTTPRequestHasAppToken(s)
	Method.LetValue(s, http.MethodPut)
	Path.Let(s, func(t *testcase.T) interface{} {
		return fmt.Sprintf(`/%s`, t.I(`id`))
	})

	s.Let(`updated-rollout`, func(t *testcase.T) interface{} {
		rf := sh.Create(release.Rollout{}).(*release.Rollout)
		rollout := sh.GetReleaseRollout(t, `rollout`)
		rf.ID = rollout.ID
		rf.FlagID = rollout.FlagID
		rf.DeploymentEnvironmentID = rollout.DeploymentEnvironmentID
		rf.Plan = *sh.Create(release.RolloutDecisionByPercentage{}).(*release.RolloutDecisionByPercentage)
		return rf
	})

	Body.Let(s, func(t *testcase.T) interface{} {
		rollout := sh.GetReleaseRollout(t, `updated-rollout`)
		var req httpapi.CreateReleaseRolloutRequest
		req.Body.Rollout.Plan = release.RolloutDefinitionView{Definition: rollout.Plan}
		return req.Body
	})

	var onSuccess = func(t *testcase.T) httpapi.UpdateReleaseRolloutResponse {
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		var resp httpapi.UpdateReleaseRolloutResponse
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.Then(`rollout is updated in the system`, func(t *testcase.T) {
		onSuccess(t)
		updatedReleaseRolloutView := *sh.GetReleaseRollout(t, `updated-rollout`)
		stored := FindStoredReleaseRollout(t)
		require.Equal(t, updatedReleaseRolloutView, stored)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			percentage := release.NewRolloutDecisionByPercentage()
			percentage.Percentage = 128
			sh.GetReleaseRollout(t, `updated-rollout`).Plan = percentage
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

			id := sh.GetReleaseRollout(t, `rollout`).ID

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			p := swagger.NewUpdateReleaseRolloutParams()
			p.RolloutID = id
			p.FlagID = sh.ExampleReleaseFlag(t).ID
			p.EnvironmentID = sh.ExampleDeploymentEnvironment(t).ID
			p.Body.Rollout = &models.Rollout{
				Plan: release.RolloutDefinitionView{Definition: release.NewRolloutDecisionByPercentage()},
			}

			resp, err := c.Release.UpdateReleaseRollout(p, protectedAuth(t))
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Payload)
		})
	})
}

func FindStoredReleaseRollout(t *testcase.T) release.Rollout {
	flag := *sh.ExampleReleaseFlag(t)
	env := *sh.ExampleDeploymentEnvironment(t)
	var r release.Rollout
	found, err := sh.StorageGet(t).FindReleaseRolloutByReleaseFlagAndDeploymentEnvironment(sh.GetContext(t), flag, env, &r)
	require.Nil(t, err)
	require.True(t, found)
	return r
}

func getPlanMap(t *testcase.T, rollout httpapi.Rollout) map[string]interface{} {
	p := rollout.Plan.(map[string]interface{})
	if s, ok := p[`seed`]; ok {
		p[`seed`] = int64(s.(float64))
	}
	if percentage, ok := p[`percentage`]; ok {
		p[`percentage`] = int(percentage.(float64))
	}
	return p
}
