package httpapi_test

import (
	"context"
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

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/external/interface/httpintf"
	"github.com/toggler-io/toggler/external/interface/httpintf/httpapi"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client"
	swagger "github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/client/deployment"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger/lib/models"
	. "github.com/toggler-io/toggler/testing"
)

func TestDeploymentEnvironmentController(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()
	SetUp(s)
	GivenThisIsAJSONAPI(s)

	LetContext(s, func(t *testcase.T) context.Context {
		return GetContext(t)
	})

	LetHandler(s, func(t *testcase.T) http.Handler {
		return httpapi.NewDeploymentEnvironmentHandler(ExampleUseCases(t))
	})

	s.Describe(`POST / - create deployment environment`, SpecDeploymentEnvironmentControllerCreate)
	s.Describe(`GET / - list deployment environment`, SpecDeploymentEnvironmentControllerList)

	s.Context(`given we have a deployment environment in the system`, func(s *testcase.Spec) {
		GivenWeHaveDeploymentEnvironment(s, `deployment-environment`)

		var andFlagIdentifierProvided = func(s *testcase.Spec, context func(s *testcase.Spec)) {
			s.And(`deployment environment identifier provided as the external ID`, func(s *testcase.Spec) {
				s.Let(`id`, func(t *testcase.T) interface{} {
					return GetDeploymentEnvironment(t, `deployment-environment`).ID
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

		andFlagIdentifierProvided(s, func(s *testcase.Spec) {
			s.Describe(`PUT|PATCH /{id} - update a deployment environment`,
				SpecDeploymentEnvironmentControllerUpdate)
		})
	})
}

func SpecDeploymentEnvironmentControllerCreate(s *testcase.Spec) {
	LetMethodValue(s, http.MethodPost)
	LetPathValue(s, `/`)
	GivenHTTPRequestHasAppToken(s)

	var onSuccess = func(t *testcase.T) (resp httpapi.CreateDeploymentEnvironmentResponse) {
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.After(func(t *testcase.T) {
		require.Nil(t, ExampleStorage(t).DeleteAll(GetContext(t), deployment.Environment{}))
	})

	s.Let(`deployment-environment`, func(t *testcase.T) interface{} {
		return FixtureFactory{}.Create(deployment.Environment{}).(*deployment.Environment)
	})

	LetBody(s, func(t *testcase.T) interface{} {
		var req httpapi.CreateDeploymentEnvironmentRequest
		req.Body.Environment = *t.I(`deployment-environment`).(*deployment.Environment)
		return req.Body
	})

	s.Then(`call succeed`, func(t *testcase.T) {
		require.Equal(t, 200, ServeHTTP(t).Code)
	})

	s.Then(`env stored in the system`, func(t *testcase.T) {
		onSuccess(t)
		rfv := t.I(`deployment-environment`).(*deployment.Environment)

		var actualDeploymentEnvironment deployment.Environment
		found, err := ExampleStorage(t).FindDeploymentEnvironmentByAlias(GetContext(t), t.I(`deployment-environment`).(*deployment.Environment).Name, &actualDeploymentEnvironment)
		require.Nil(t, err)
		require.True(t, found)
		require.Equal(t, rfv.Name, actualDeploymentEnvironment.Name)
	})

	s.Then(`it returns env in the response`, func(t *testcase.T) {
		resp := onSuccess(t)

		var env deployment.Environment
		found, err := ExampleStorage(t).FindDeploymentEnvironmentByAlias(GetContext(t), t.I(`deployment-environment`).(*deployment.Environment).Name, &env)
		require.Nil(t, err)
		require.True(t, found)
		require.Equal(t, resp.Body.Environment, env)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`for example name is empty`)
			t.I(`deployment-environment`).(*deployment.Environment).Name = ``
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
		s.Tag(TagBlackBox)

		s.Test(`swagger`, func(t *testcase.T) {
			sm, err := httpintf.NewServeMux(ExampleUseCases(t))
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
	LetMethodValue(s, http.MethodGet)
	LetPathValue(s, `/`)
	GivenHTTPRequestHasAppToken(s)

	var onSuccess = func(t *testcase.T) httpapi.ListDeploymentEnvironmentResponse {
		var resp httpapi.ListDeploymentEnvironmentResponse
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.And(`no flag present in the system`, func(s *testcase.Spec) {
		NoDeploymentEnvironmentPresentInTheStorage(s)

		s.Then(`empty result received`, func(t *testcase.T) {
			require.Empty(t, onSuccess(t).Body.Environments)
		})
	})

	s.And(`deployment environment is present in the system`, func(s *testcase.Spec) {
		GivenWeHaveDeploymentEnvironment(s, `feature-1`)
		s.Before(func(t *testcase.T) { GetDeploymentEnvironment(t, `feature-1`) }) // eager load

		s.Then(`env received back`, func(t *testcase.T) {
			resp := onSuccess(t)
			require.Len(t, resp.Body.Environments, 1)
			require.Contains(t, resp.Body.Environments, *GetDeploymentEnvironment(t, `feature-1`))
		})

		s.And(`even multiple flag in the system`, func(s *testcase.Spec) {
			GivenWeHaveDeploymentEnvironment(s, `feature-2`)
			s.Before(func(t *testcase.T) { GetDeploymentEnvironment(t, `feature-2`) }) // eager load

			s.Then(`the flags will be received back`, func(t *testcase.T) {
				resp := onSuccess(t)

				require.Len(t, resp.Body.Environments, 2)
				require.Contains(t, resp.Body.Environments, *GetDeploymentEnvironment(t, `feature-2`))
			})
		})
	})
}

func SpecDeploymentEnvironmentControllerUpdate(s *testcase.Spec) {
	GivenHTTPRequestHasAppToken(s)
	LetMethodValue(s, http.MethodPut)
	LetPath(s, func(t *testcase.T) string {
		return fmt.Sprintf(`/%s`, t.I(`id`))
	})

	s.Let(`updated-deployment-environment`, func(t *testcase.T) interface{} {
		rf := FixtureFactory{}.Create(deployment.Environment{}).(*deployment.Environment)
		rf.ID = GetDeploymentEnvironment(t, `deployment-environment`).ID
		return rf
	})

	LetBody(s, func(t *testcase.T) interface{} {
		var req httpapi.CreateDeploymentEnvironmentRequest
		req.Body.Environment = *t.I(`updated-deployment-environment`).(*deployment.Environment)
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

		updatedDeploymentEnvironmentView := t.I(`updated-deployment-environment`).(*deployment.Environment)

		var stored deployment.Environment
		found, err := ExampleStorage(t).FindDeploymentEnvironmentByAlias(GetContext(t), updatedDeploymentEnvironmentView.Name, &stored)
		require.Nil(t, err)
		require.True(t, found)
		require.Equal(t, resp.Body.Environment, stored)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`for example the name is empty`)
			t.I(`updated-deployment-environment`).(*deployment.Environment).Name = ``
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
		s.Tag(TagBlackBox)

		s.Test(`swagger`, func(t *testcase.T) {
			sm, err := httpintf.NewServeMux(ExampleUseCases(t))
			require.Nil(t, err)

			s := httptest.NewServer(sm)
			defer s.Close()

			id := GetDeploymentEnvironment(t, `deployment-environment`).ID

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
