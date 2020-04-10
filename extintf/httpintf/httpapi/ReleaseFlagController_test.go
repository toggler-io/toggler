package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/gorest"
	. "github.com/adamluzsi/testcase/httpspec"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/extintf/httpintf"
	"github.com/toggler-io/toggler/extintf/httpintf/httpapi"
	"github.com/toggler-io/toggler/lib/go/client"
	swagger "github.com/toggler-io/toggler/lib/go/client/release"
	"github.com/toggler-io/toggler/lib/go/models"
	"github.com/toggler-io/toggler/usecases"

	. "github.com/toggler-io/toggler/testing"
)

func TestReleaseFlagController(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()
	GivenThisIsAJSONAPI(s)
	SetUp(s)

	LetHandler(s, func(t *testcase.T) http.Handler {
		return gorest.NewHandler(httpapi.ReleaseFlagController{UseCases: ExampleUseCases(t)})
	})

	s.Describe(`POST / - create release flag`, func(s *testcase.Spec) {
		LetMethodValue(s, http.MethodPost)
		LetPathValue(s, `/`)

		var onSuccess = func(t *testcase.T) (resp httpapi.CreateReleaseFlagResponse) {
			rr := ServeHTTP(t)
			require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
			require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
			return resp
		}

		s.Let(`release-flag`, func(t *testcase.T) interface{} {
			return FixtureFactory{}.Create(release.Flag{}).(*release.Flag)
		})

		LetBody(s, func(t *testcase.T) interface{} {
			var req httpapi.CreateReleaseFlagRequest
			req.Body.Flag = *t.I(`release-flag`).(*release.Flag)
			return req.Body
		})

		s.Then(`call succeed`, func(t *testcase.T) {
			require.Equal(t, 200, ServeHTTP(t).Code)
		})

		s.Then(`flag stored in the system`, func(t *testcase.T) {
			onSuccess(t)
			expectedReleaseFlag := t.I(`release-flag`).(*release.Flag)
			actualReleaseFlag := FindStoredReleaseFlagByName(t, expectedReleaseFlag.Name)
			actualReleaseFlag.ID = `` // because the provided release flag has no id
			require.Equal(t, expectedReleaseFlag, actualReleaseFlag)
		})

		s.Then(`it returns flag in the response`, func(t *testcase.T) {
			resp := onSuccess(t)
			releaseFlag := *FindStoredReleaseFlagByName(t, t.I(`release-flag`).(*release.Flag).Name)
			require.Equal(t, releaseFlag, resp.Body.Flag)
		})

		s.And(`if input contains invalid values`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				t.Log(`for example the rollout percentage is invalid`)
				t.I(`release-flag`).(*release.Flag).Rollout.Strategy.Percentage = 150
			})

			s.Then(`it will return with failure`, func(t *testcase.T) {
				rr := ServeHTTP(t)
				require.Equal(t, http.StatusBadRequest, rr.Code)

				var resp httpapi.ErrorResponse
				require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body), rr.Body.String())
				require.Equal(t, release.ErrInvalidPercentage.Error(), resp.Body.Error.Message)
			})
		})

		s.Test(`swagger`, func(t *testcase.T) {
			sm, err := httpintf.NewServeMux(usecases.NewUseCases(ExampleStorage(t)))
			require.Nil(t, err)

			s := httptest.NewServer(sm)
			defer s.Close()

			// TODO: ensure validation
			p := swagger.NewCreateReleaseFlagParams()
			p.Body.Flag = &models.Flag{
				Name: fixtures.Random.String(),
				Rollout: &models.FlagRollout{
					Strategy: &models.FlagRolloutStrategy{
						Percentage: int64(fixtures.Random.IntBetween(0, 100)),
					},
				},
			}

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			resp, err := c.Release.CreateReleaseFlag(p)
			if err != nil {
				t.Fatal(err.Error())
			}

			require.NotNil(t, resp)
			require.NotNil(t, resp.Payload)

		})
	})
}
