package httpapi_test

import (
	"encoding/json"
	"fmt"
	"github.com/toggler-io/toggler/domains/release"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/testcase/httpspec"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/external/interface/httpintf"
	"github.com/toggler-io/toggler/external/interface/httpintf/httpapi"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client"
	swagger "github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client/deployment"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/models"
	sh "github.com/toggler-io/toggler/spechelper"
)

var env = testcase.Var{
	Name: `deployment-environment`,
	Init: func(t *testcase.T) interface{} {
		e := sh.NewFixtureFactory(t).Create(release.Environment{}).(release.Environment)
		return &e
	},
}

func envGet(t *testcase.T) *release.Environment {
	return env.Get(t).(*release.Environment)
}

func TestDeploymentEnvironmentController(t *testing.T) {
	s := sh.NewSpec(t)
	s.Parallel()

	Handler.Let(s, func(t *testcase.T) interface{} {
		return httpapi.NewDeploymentEnvironmentHandler(sh.ExampleUseCases(t))
	})

	ContentTypeIsJSON(s)

	Context.Let(s, func(t *testcase.T) interface{} {
		return sh.ContextGet(t)
	})

	s.Describe(`POST / - create deployment environment`, SpecDeploymentEnvironmentControllerCreate)
	s.Describe(`GET / - list deployment environment`, SpecDeploymentEnvironmentControllerList)

	s.Context(`given we have a deployment environment in the system`, func(s *testcase.Spec) {
		sh.GivenWeHaveDeploymentEnvironment(s, env.Name)

		var andDeploymentIdentifierProvided = func(s *testcase.Spec, context func(s *testcase.Spec)) {
			s.And(`deployment environment identifier provided as the external ID`, func(s *testcase.Spec) {
				s.Let(`id`, func(t *testcase.T) interface{} {
					return envGet(t).ID
				})

				context(s)
			})

			//s.And(`deployment environment identifier provided as url normalized deployment environment name`, func(s *testcase.Spec) {
			//	// TODO add implementation to "alias id that can be guessed from the flag name"
			//
			//	s.Let(`id`, func(t *testcase.T) interface{} {
			//		return GetDeploymentEnvironment(t, `deployment-environment`).Name
			//	})
			//
			//	context(s)
			//})
		}

		andDeploymentIdentifierProvided(s, func(s *testcase.Spec) {
			s.Describe(`PUT|PATCH /{id} - update a deployment environment`,
				SpecDeploymentEnvironmentControllerUpdate)

			s.Describe(`DELETE /{id} - delete a deployment environment`,
				SpecDeploymentEnvironmentControllerDelete)
		})
	})
}

func SpecDeploymentEnvironmentControllerCreate(s *testcase.Spec) {
	Method.LetValue(s, http.MethodPost)
	Path.LetValue(s, `/`)
	sh.GivenHTTPRequestHasAppToken(s)

	var onSuccess = func(t *testcase.T) (resp httpapi.CreateDeploymentEnvironmentResponse) {
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.After(func(t *testcase.T) {
		require.Nil(t, sh.StorageGet(t).ReleaseEnvironment(sh.ContextGet(t)).DeleteAll(sh.ContextGet(t)))
	})

	Body.Let(s, func(t *testcase.T) interface{} {
		var req httpapi.CreateDeploymentEnvironmentRequest
		req.Body.Environment = *envGet(t)
		return req.Body
	})

	s.Then(`call succeed`, func(t *testcase.T) {
		require.Equal(t, 200, ServeHTTP(t).Code)
	})

	s.Then(`env stored in the system`, func(t *testcase.T) {
		onSuccess(t)
		rfv := envGet(t)

		var actualDeploymentEnvironment release.Environment
		found, err := sh.StorageGet(t).ReleaseEnvironment(sh.ContextGet(t)).FindByAlias(sh.ContextGet(t), envGet(t).Name, &actualDeploymentEnvironment)
		require.Nil(t, err)
		require.True(t, found)
		require.Equal(t, rfv.Name, actualDeploymentEnvironment.Name)
	})

	s.Then(`it returns env in the response`, func(t *testcase.T) {
		resp := onSuccess(t)

		var env release.Environment
		found, err := sh.StorageGet(t).ReleaseEnvironment(sh.ContextGet(t)).FindByAlias(sh.ContextGet(t), envGet(t).Name, &env)
		require.Nil(t, err)
		require.True(t, found)
		require.Equal(t, resp.Body.Environment, env)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`for example name is empty`)
			envGet(t).Name = ``
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
			p := swagger.NewCreateDeploymentEnvironmentParams()
			p.Body.Environment = &models.Environment{
				Name: fixtures.Random.String(),
			}

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			resp, err := c.Deployment.CreateDeploymentEnvironment(p, protectedAuth(t))
			require.Nil(t, err)

			require.NotNil(t, resp)
			require.NotNil(t, resp.Payload)
		})
	})
}

func SpecDeploymentEnvironmentControllerList(s *testcase.Spec) {
	Method.LetValue(s, http.MethodGet)
	Path.LetValue(s, `/`)
	sh.GivenHTTPRequestHasAppToken(s)

	var onSuccess = func(t *testcase.T) httpapi.ListDeploymentEnvironmentResponse {
		var resp httpapi.ListDeploymentEnvironmentResponse
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.And(`no flag present in the system`, func(s *testcase.Spec) {
		sh.NoDeploymentEnvironmentPresentInTheStorage(s)

		s.Then(`empty result received`, func(t *testcase.T) {
			require.Empty(t, onSuccess(t).Body.Environments)
		})
	})

	s.And(`deployment environment is present in the system`, func(s *testcase.Spec) {
		sh.GivenWeHaveDeploymentEnvironment(s, `feature-1`).EagerLoading(s)

		s.Then(`env received back`, func(t *testcase.T) {
			resp := onSuccess(t)
			require.Len(t, resp.Body.Environments, 1)
			require.Contains(t, resp.Body.Environments, *sh.GetDeploymentEnvironment(t, `feature-1`))
		})

		s.And(`even multiple flag in the system`, func(s *testcase.Spec) {
			sh.GivenWeHaveDeploymentEnvironment(s, `feature-2`).EagerLoading(s)

			s.Then(`the flags will be received back`, func(t *testcase.T) {
				resp := onSuccess(t)

				require.Len(t, resp.Body.Environments, 2)
				require.Contains(t, resp.Body.Environments, *sh.GetDeploymentEnvironment(t, `feature-2`))
			})
		})
	})
}

func SpecDeploymentEnvironmentControllerUpdate(s *testcase.Spec) {
	sh.GivenHTTPRequestHasAppToken(s)
	Method.LetValue(s, http.MethodPut)

	Path.Let(s, func(t *testcase.T) interface{} {
		return fmt.Sprintf(`/%s`, t.I(`id`))
	})

	updatedEnv := s.Let(`updated-deployment-environment`, func(t *testcase.T) interface{} {
		rf := sh.NewFixtureFactory(t).Create(release.Environment{}).(release.Environment)
		rf.ID = envGet(t).ID
		return &rf
	})
	updatedEnvGet := func(t *testcase.T) *release.Environment {
		return updatedEnv.Get(t).(*release.Environment)
	}

	Body.Let(s, func(t *testcase.T) interface{} {
		var req httpapi.CreateDeploymentEnvironmentRequest
		req.Body.Environment = *updatedEnvGet(t)
		return req.Body
	})

	var onSuccess = func(t *testcase.T) httpapi.UpdateDeploymentEnvironmentResponse {
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code)
		var resp httpapi.UpdateDeploymentEnvironmentResponse
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.Then(`env is updated in the system`, func(t *testcase.T) {
		resp := onSuccess(t)

		updatedDeploymentEnvironmentView := updatedEnvGet(t)

		var stored release.Environment
		found, err := sh.StorageGet(t).ReleaseEnvironment(sh.ContextGet(t)).FindByAlias(sh.ContextGet(t), updatedDeploymentEnvironmentView.Name, &stored)
		require.Nil(t, err)
		require.True(t, found)
		require.Equal(t, resp.Body.Environment, stored)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`for example the name is empty`)
			updatedEnvGet(t).Name = ``
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

			id := envGet(t).ID

			// TODO: ensure validation
			p := swagger.NewUpdateDeploymentEnvironmentParams()
			p.EnvironmentID = id
			p.Body.Environment = &models.Environment{
				ID:   id,
				Name: fixtures.Random.String(),
			}

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			resp, err := c.Deployment.UpdateDeploymentEnvironment(p, protectedAuth(t))
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Payload)
		})
	})
}

func SpecDeploymentEnvironmentControllerDelete(s *testcase.Spec) {
	sh.GivenHTTPRequestHasAppToken(s)
	Method.LetValue(s, http.MethodDelete)

	Path.Let(s, func(t *testcase.T) interface{} {
		return fmt.Sprintf(`/%s`, t.I(`id`))
	})

	var onSuccess = func(t *testcase.T) {
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code)
	}

	s.Then(`env is deleted from the system`, func(t *testcase.T) {
		onSuccess(t)

		deletedDeploymentEnvironment := envGet(t)

		var stored release.Environment
		found, err := sh.StorageGet(t).ReleaseEnvironment(sh.ContextGet(t)).FindByAlias(sh.ContextGet(t), deletedDeploymentEnvironment.Name, &stored)
		require.Nil(t, err)
		require.False(t, found)
		require.Equal(t, release.Environment{}, stored)
	})

	s.Context(`E2E`, func(s *testcase.Spec) {
		s.Tag(sh.TagBlackBox)

		s.Test(`swagger`, func(t *testcase.T) {
			sm, err := httpintf.NewServeMux(sh.ExampleUseCases(t))
			require.Nil(t, err)

			s := httptest.NewServer(sm)
			defer s.Close()

			id := envGet(t).ID

			// TODO: ensure validation
			p := swagger.NewDeleteDeploymentEnvironmentParams()
			p.EnvironmentID = id

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			_, err = c.Deployment.DeleteDeploymentEnvironment(p, protectedAuth(t))
			require.Nil(t, err)
		})
	})
}
