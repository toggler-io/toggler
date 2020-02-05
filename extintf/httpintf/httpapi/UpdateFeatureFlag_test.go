package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	. "github.com/toggler-io/toggler/testing"
)

func TestServeMux_UpdateFeatureFlag(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	subject := func(t *testcase.T) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		NewServeMux(t).ServeHTTP(rr, t.I(`request`).(*http.Request))
		return rr
	}

	SetupSpecCommonVariables(s)

	s.Let(`enrollment query value`, func(t *testcase.T) interface{} {
		return strconv.FormatBool(GetPilotEnrollment(t))
	})

	s.Let(`TokenString`, func(t *testcase.T) interface{} {
		return GetTextToken(t)
	})

	s.Let(`request`, func(t *testcase.T) interface{} {
		u, err := url.Parse(`/release/flag` + t.I(`http path`).(string))
		require.Nil(t, err)

		values := u.Query()
		values.Set(`token`, t.I(`TokenString`).(string))
		u.RawQuery = values.Encode()

		payload := bytes.NewBuffer(t.I(`payload bytes`).([]byte))
		req := httptest.NewRequest(http.MethodPost, u.String(), payload)

		req.Header.Set(`Content-Type`, t.I(`Content-Type`).(string))

		return req
	})

	s.Before(func(t *testcase.T) {
		t.Log(`given we have flag already stored`)
		require.Nil(t, GetStorage(t).Create(context.TODO(), GetReleaseFlag(t)))
	})

	s.When(`request is sent to the JSON endpoint`, func(s *testcase.Spec) {

		s.Let(`Content-Type`, func(t *testcase.T) interface{} {
			return `application/json`
		})

		s.Let(`payload bytes`, func(t *testcase.T) interface{} {
			bs, err := json.Marshal(GetReleaseFlag(t))
			require.Nil(t, err)
			return bs
		})

		s.Let(`http path`, func(t *testcase.T) interface{} {
			return `/update.json`
		})

		s.Then(`it will reply back in json format`, func(t *testcase.T) {
			var resp struct{}
			r := subject(t)
			IsJsonResponse(t, r, &resp)
		})

		SpecServeMux_UpdateFeatureFlag(s, subject)

	})

	s.When(`request is send as x-www-form-urlencoded payload`, func(s *testcase.Spec) {

		s.Let(`Content-Type`, func(t *testcase.T) interface{} {
			return `application/x-www-form-urlencoded`
		})

		s.Let(`payload bytes`, func(t *testcase.T) interface{} {
			data := url.Values{}
			data.Set(`flag.id`, GetReleaseFlag(t).ID)
			data.Set(`flag.name`, GetReleaseFlag(t).Name)
			data.Set(`flag.rollout.randSeed`, strconv.FormatInt(GetReleaseFlag(t).Rollout.RandSeed, 10))
			data.Set(`flag.rollout.strategy.percentage`, strconv.Itoa(GetReleaseFlag(t).Rollout.Strategy.Percentage))

			var decisionLogicAPI string
			if GetReleaseFlag(t).Rollout.Strategy.DecisionLogicAPI != nil {
				decisionLogicAPI = GetReleaseFlag(t).Rollout.Strategy.DecisionLogicAPI.String()
			}
			data.Set(`flag.rollout.strategy.decisionLogicApi`, decisionLogicAPI)

			return []byte(data.Encode())
		})

		s.Let(`http path`, func(t *testcase.T) interface{} {
			return `/update.form`
		})

		s.And(`the decision api url`, func(s *testcase.Spec) {
			s.Context(`is a valid url`, func(s *testcase.Spec) {
				s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} {
					return `http://mydomain/api/experiment`
				})

				// then it will persist the url
				SpecServeMux_UpdateFeatureFlag(s, subject)
			})

			s.Context(`is an invalid url`, func(s *testcase.Spec) {
				s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} {
					return `hello world!`
				})

				//TODO
			})

			s.Context(`is empty`, func(s *testcase.Spec) {
				s.Let(`RolloutApiURL`, func(t *testcase.T) interface{} { return nil })

				s.Context(`the flag decision api url will be ereased`, func(s *testcase.Spec) {
					SpecServeMux_UpdateFeatureFlag(s, subject)
				})
			})
		})

		s.And(`form flag rollout strategy percentage`, func(s *testcase.Spec) {
			s.Context(`is a valid`, func(s *testcase.Spec) {
				s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return 42 })

				// then it will persist the url
				SpecServeMux_UpdateFeatureFlag(s, subject)
			})

			s.Context(`is an invalid url`, func(s *testcase.Spec) {
				//TODO
			})

			s.Context(`is empty`, func(s *testcase.Spec) {
				//TODO
			})
		})

	})

}

func SpecServeMux_UpdateFeatureFlag(s *testcase.Spec, subject func(t *testcase.T) *httptest.ResponseRecorder) {
	s.And(`invalid token given`, func(s *testcase.Spec) {
		s.Let(`TokenString`, func(t *testcase.T) interface{} {
			return `invalid`
		})

		s.Then(`it will return unauthorized`, func(t *testcase.T) {
			r := subject(t)

			require.Equal(t, 401, r.Code)
		})
	})

	s.And(`valid token provided`, func(s *testcase.Spec) {
		s.Let(`TokenString`, func(t *testcase.T) interface{} {
			return GetTextToken(t)
		})

		s.Then(`call succeed`, func(t *testcase.T) {
			r := subject(t)
			require.Equal(t, 200, r.Code, r.Body.String())
		})

		s.Then(`flag stored in the system`, func(t *testcase.T) {
			r := subject(t)
			require.Equal(t, 200, r.Code, r.Body.String())

			stored := FindStoredReleaseFlagByName(t)

			require.Equal(t, GetReleaseFlag(t), stored)
		})
	})
}
