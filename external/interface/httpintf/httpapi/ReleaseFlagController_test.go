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
	"github.com/toggler-io/toggler/lib/go/client"
	swagger "github.com/toggler-io/toggler/lib/go/client/release"
	"github.com/toggler-io/toggler/lib/go/models"
	. "github.com/toggler-io/toggler/testing"
)

func TestReleaseFlagView(t *testing.T) {
	s := testcase.NewSpec(t)
	s.NoSideEffect()

	s.Describe(`FromReleaseFlag`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) httpapi.ReleaseFlagView {
			return httpapi.ReleaseFlagView{}.FromReleaseFlag(*t.I(`release-flag`).(*release.Flag))
		}

		s.Let(`release-flag`, func(t *testcase.T) interface{} {
			flag := FixtureFactory{}.Create(release.Flag{}).(*release.Flag)
			flag.ID = `42`
			return flag
		})

		s.Then(`it will map release flag values into the view model`, func(t *testcase.T) {
			flag := t.I(`release-flag`).(*release.Flag)
			view := subject(t)
			require.Equal(t, view.ID, flag.ID)
			require.Equal(t, view.Name, flag.Name)
			require.Equal(t, view.Rollout.RandSeed, flag.Rollout.RandSeed)
			require.Equal(t, view.Rollout.Strategy.Percentage, flag.Rollout.Strategy.Percentage)
			require.Equal(t, view.Rollout.Strategy.DecisionLogicAPI, flag.Rollout.Strategy.DecisionLogicAPI)
		})
	})

	s.Describe(`ToReleaseFlag`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) release.Flag {
			return t.I(`view`).(httpapi.ReleaseFlagView).ToReleaseFlag()
		}
		s.Let(`view`, func(t *testcase.T) interface{} {
			return httpapi.ReleaseFlagView{}.FromReleaseFlag(*t.I(`release-flag`).(*release.Flag))
		})
		s.Let(`release-flag`, func(t *testcase.T) interface{} {
			flag := FixtureFactory{}.Create(release.Flag{}).(*release.Flag)
			flag.ID = `42`
			return flag
		})

		s.Then(`it will convert the view into a release flag`, func(t *testcase.T) {
			require.Equal(t, *t.I(`release-flag`).(*release.Flag), subject(t))
		})
	})

	s.Describe(`json keys in serialization`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) string {
			var rfv httpapi.ReleaseFlagView
			flag := FixtureFactory{}.Create(release.Flag{}).(*release.Flag)
			rfv = rfv.FromReleaseFlag(*flag)
			rfv.ID = `42`
			bs, err := json.Marshal(rfv)
			require.Nil(t, err)
			return string(bs)
		}

		s.Test(`id`, func(t *testcase.T) {
			require.Contains(t, subject(t), `"id":`)
		})

		s.Test(`name`, func(t *testcase.T) {
			require.Contains(t, subject(t), `"name":`)
		})

		s.Test(`rollout`, func(t *testcase.T) {
			require.Contains(t, subject(t), `"rollout":`)
		})

		s.Test(`strategy`, func(t *testcase.T) {
			require.Contains(t, subject(t), `"strategy":`)
		})

		s.Test(`percentage`, func(t *testcase.T) {
			require.Contains(t, subject(t), `"percentage":`)
		})

		s.Test(`decision_logic_api`, func(t *testcase.T) {
			require.Contains(t, subject(t), `"decision_logic_api":`)
		})

		s.Test(`rand_seed_salt`, func(t *testcase.T) {
			require.Contains(t, subject(t), `"rand_seed_salt":`)
		})
	})
}

func TestReleaseFlagController(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()
	SetUp(s)
	GivenThisIsAJSONAPI(s)

	LetHandler(s, func(t *testcase.T) http.Handler {
		return httpapi.NewReleaseFlagHandler(ExampleUseCases(t))
	})

	s.Describe(`POST / - create release flag`, SpecReleaseFlagControllerCreate)
	s.Describe(`GET / - list release flags`, SpecReleaseFlagControllerList)

	s.Context(`given we have a release flag in the system`, func(s *testcase.Spec) {
		GivenWeHaveReleaseFlag(s, `release-flag`)

		var andFlagIdentifierProvided = func(s *testcase.Spec, context func(s *testcase.Spec)) {
			s.And(`release flag identifier provided as the external ID`, func(s *testcase.Spec) {
				s.Let(`id`, func(t *testcase.T) interface{} {
					return GetReleaseFlag(t, `release-flag`).ID
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
			s.Describe(`GET /{id||alias}/global - get release flag's global state`,
				SpecReleaseFlagControllerReleaseFlagGlobalStats)

			s.Describe(`PUT|PATCH /{id} - update a release flag`,
				SpecReleaseFlagControllerUpdate)
		})
	})
}

func SpecReleaseFlagControllerCreate(s *testcase.Spec) {
	LetMethodValue(s, http.MethodPost)
	LetPathValue(s, `/`)
	GivenHTTPRequestHasAppToken(s)

	var onSuccess = func(t *testcase.T) (resp httpapi.CreateReleaseFlagResponse) {
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.Let(`release-flag`, func(t *testcase.T) interface{} {
		rf := FixtureFactory{}.Create(release.Flag{}).(*release.Flag)
		rfv := httpapi.ReleaseFlagView{}.FromReleaseFlag(*rf)
		return &rfv
	})

	LetBody(s, func(t *testcase.T) interface{} {
		var req httpapi.CreateReleaseFlagRequest
		req.Body.Flag = *t.I(`release-flag`).(*httpapi.ReleaseFlagView)
		return req.Body
	})

	s.Then(`call succeed`, func(t *testcase.T) {
		require.Equal(t, 200, ServeHTTP(t).Code)
	})

	s.Then(`flag stored in the system`, func(t *testcase.T) {
		onSuccess(t)
		rfv := t.I(`release-flag`).(*httpapi.ReleaseFlagView)
		actualReleaseFlag := FindStoredReleaseFlagByName(t, rfv.Name)
		require.Equal(t, rfv.Name, actualReleaseFlag.Name)
		require.Equal(t, rfv.Rollout.RandSeed, actualReleaseFlag.Rollout.RandSeed)
		require.Equal(t, rfv.Rollout.Strategy.Percentage, actualReleaseFlag.Rollout.Strategy.Percentage)
		require.Equal(t, rfv.Rollout.Strategy.DecisionLogicAPI, actualReleaseFlag.Rollout.Strategy.DecisionLogicAPI)
	})

	s.Then(`it returns flag in the response`, func(t *testcase.T) {
		resp := onSuccess(t)
		flag := *FindStoredReleaseFlagByName(t, t.I(`release-flag`).(*httpapi.ReleaseFlagView).Name)
		require.Equal(t, flag.Name, resp.Body.Flag.Name)
		require.Equal(t, flag.Rollout.RandSeed, resp.Body.Flag.Rollout.RandSeed)
		require.Equal(t, flag.Rollout.Strategy.Percentage, resp.Body.Flag.Rollout.Strategy.Percentage)
		require.Equal(t, flag.Rollout.Strategy.DecisionLogicAPI, resp.Body.Flag.Rollout.Strategy.DecisionLogicAPI)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`for example the rollout percentage is invalid`)
			t.I(`release-flag`).(*httpapi.ReleaseFlagView).Rollout.Strategy.Percentage = 150
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
		sm, err := httpintf.NewServeMux(ExampleUseCases(t))
		require.Nil(t, err)

		s := httptest.NewServer(sm)
		defer s.Close()

		// TODO: ensure validation
		p := swagger.NewCreateReleaseFlagParams()
		p.Body.Flag = &models.ReleaseFlagView{
			Name: fixtures.Random.String(),
			Rollout: &models.Rollout{
				Strategy: &models.Strategy{
					Percentage: int64(fixtures.Random.IntBetween(0, 100)),
				},
			},
		}

		tc := client.DefaultTransportConfig()
		u, _ := url.Parse(s.URL)
		tc.Host = u.Host
		tc.Schemes = []string{`http`}

		c := client.NewHTTPClientWithConfig(nil, tc)

		resp, err := c.Release.CreateReleaseFlag(p, protectedAuth(t))
		if err != nil {
			t.Fatal(err.Error())
		}

		require.NotNil(t, resp)
		require.NotNil(t, resp.Payload)
	})
}

func SpecReleaseFlagControllerList(s *testcase.Spec) {
	LetMethodValue(s, http.MethodGet)
	LetPathValue(s, `/`)
	GivenHTTPRequestHasAppToken(s)

	var onSuccess = func(t *testcase.T) httpapi.ListReleaseFlagResponse {
		var resp httpapi.ListReleaseFlagResponse
		rr := ServeHTTP(t)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.And(`no flag present in the system`, func(s *testcase.Spec) {
		NoReleaseFlagPresentInTheStorage(s)

		s.Then(`empty result received`, func(t *testcase.T) {
			require.Empty(t, onSuccess(t).Body.Flags)
		})
	})

	s.And(`feature flag is present in the system`, func(s *testcase.Spec) {
		GivenWeHaveReleaseFlag(s, `feature-1`)
		s.Before(func(t *testcase.T) { t.I(`feature-1`) }) // eager load

		s.Then(`flag(s) received back`, func(t *testcase.T) {
			resp := onSuccess(t)
			require.Len(t, resp.Body.Flags, 1)
			require.Equal(t, resp.Body.Flags[0].Name, t.I(`feature-1`).(*release.Flag).Name)
		})

		s.And(`even multiple flag in the system`, func(s *testcase.Spec) {
			GivenWeHaveReleaseFlag(s, `feature-2`)
			s.Before(func(t *testcase.T) { t.I(`feature-2`) }) // eager load

			s.Then(`the flags will be received back`, func(t *testcase.T) {
				resp := onSuccess(t)

				require.Len(t, resp.Body.Flags, 2)
				require.Contains(t, resp.Body.Flags, httpapi.ReleaseFlagView{}.FromReleaseFlag(*t.I(`feature-2`).(*release.Flag)))
			})
		})
	})
}

func SpecReleaseFlagControllerReleaseFlagGlobalStats(s *testcase.Spec) {
	LetMethodValue(s, http.MethodGet)
	LetPath(s, func(t *testcase.T) string {
		return fmt.Sprintf(`/%s/global`, t.I(`id`))
	})

	var onSuccess = func(t *testcase.T) httpapi.GetReleaseFlagGlobalStatesResponse {
		rr := ServeHTTP(t)
		require.Equal(t, 200, rr.Code, rr.Body.String())
		var resp httpapi.GetReleaseFlagGlobalStatesResponse
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp.Body))
		return resp
	}

	s.And(`flag global released`, func(s *testcase.Spec) {
		AndReleaseFlagPercentageIs(s, `release-flag`, 100)

		s.Then(`the request will be accepted with OK`, func(t *testcase.T) {
			require.Equal(t, true, onSuccess(t).Body.Enrollment)
		})
	})

	s.And(`flag is not fully released yet`, func(s *testcase.Spec) {
		AndReleaseFlagPercentageIs(s, `release-flag`, fixtures.Random.IntN(100))

		s.Then(`the flag enrollment will be marked as forbidden`, func(t *testcase.T) {
			require.Equal(t, false, onSuccess(t).Body.Enrollment)
		})
	})

	s.Context(`swagger`, func(s *testcase.Spec) {
		AndReleaseFlagPercentageIs(s, `release-flag`, 100)

		s.Test(`E2E`, func(t *testcase.T) {
			s := httptest.NewServer(NewServeMux(t))
			defer s.Close()

			p := swagger.NewGetReleaseFlagGlobalStatesParams()
			p.FlagID = GetReleaseFlag(t, `release-flag`).ID

			tc := client.DefaultTransportConfig()
			u, _ := url.Parse(s.URL)
			tc.Host = u.Host
			tc.Schemes = []string{`http`}

			c := client.NewHTTPClientWithConfig(nil, tc)

			resp, err := c.Release.GetReleaseFlagGlobalStates(p, publicAuth(t))
			t.Log(resp, err)
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Payload)
			require.True(t, resp.Payload.Enrollment)
		})
	})
}

func SpecReleaseFlagControllerUpdate(s *testcase.Spec) {
	GivenHTTPRequestHasAppToken(s)
	LetMethodValue(s, http.MethodPut)
	LetPath(s, func(t *testcase.T) string {
		return fmt.Sprintf(`/%s`, t.I(`id`))
	})

	s.Let(`updated-release-flag`, func(t *testcase.T) interface{} {
		rf := FixtureFactory{}.Create(release.Flag{}).(*release.Flag)
		rfv := httpapi.ReleaseFlagView{}.FromReleaseFlag(*rf)
		rfv.ID = GetReleaseFlag(t, `release-flag`).ID
		return &rfv
	})

	LetBody(s, func(t *testcase.T) interface{} {
		var req httpapi.CreateReleaseFlagRequest
		req.Body.Flag = *t.I(`updated-release-flag`).(*httpapi.ReleaseFlagView)
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

		updatedReleaseFlagView := t.I(`updated-release-flag`).(*httpapi.ReleaseFlagView)
		stored := FindStoredReleaseFlagByName(t, updatedReleaseFlagView.Name)
		require.Equal(t, resp.Body.Flag.ID, stored.ID)
		require.Equal(t, resp.Body.Flag.ToReleaseFlag(), *stored)
	})

	s.And(`if input contains invalid values`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`for example the rollout percentage is invalid`)
			t.I(`updated-release-flag`).(*httpapi.ReleaseFlagView).Rollout.Strategy.Percentage = 150
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
		sm, err := httpintf.NewServeMux(ExampleUseCases(t))
		require.Nil(t, err)

		s := httptest.NewServer(sm)
		defer s.Close()

		id := GetReleaseFlag(t, `release-flag`).ID

		// TODO: ensure validation
		p := swagger.NewUpdateReleaseFlagParams()
		p.FlagID = id
		p.Body.Flag = &models.ReleaseFlagView{
			ID:   id,
			Name: fixtures.Random.String(),
			Rollout: &models.Rollout{
				Strategy: &models.Strategy{
					Percentage: int64(fixtures.Random.IntBetween(0, 100)),
				},
			},
		}

		tc := client.DefaultTransportConfig()
		u, _ := url.Parse(s.URL)
		tc.Host = u.Host
		tc.Schemes = []string{`http`}

		c := client.NewHTTPClientWithConfig(nil, tc)

		resp, err := c.Release.UpdateReleaseFlag(p, protectedAuth(t))
		require.Nil(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Payload)
	})
}
