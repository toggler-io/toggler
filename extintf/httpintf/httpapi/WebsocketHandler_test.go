package httpapi_test

import (
	"context"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/toggler/extintf/httpintf/httpapi"
	. "github.com/adamluzsi/toggler/testing"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebsocket(t *testing.T) {
	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)

	server := func(t *testcase.T) *httptest.Server { return t.I(`server`).(*httptest.Server) }
	s.Let(`server`, func(t *testcase.T) interface{} {
		return httptest.NewServer(withAccessLogging(t, NewServeMux(t)))
	})
	s.After(func(t *testcase.T) { server(t).Close() })

	s.Let(`url`, func(t *testcase.T) interface{} {
		url := server(t).URL
		url = "ws" + strings.TrimPrefix(url, "http")
		url += `/ws`
		return url
	})

	ws := func(t *testcase.T) *websocket.Conn { return t.I(`ws`).(*websocket.Conn) }
	s.Let(`ws`, func(t *testcase.T) interface{} {
		url := t.I(`url`).(string)
		ws, resp, err := websocket.DefaultDialer.Dial(url, nil)
		require.NotNil(t, resp)
		if err == websocket.ErrBadHandshake && resp != nil && err != nil {
			t.Logf(`target url is %q`, url)
			t.Fatalf(`%s with HTTP status code %d`, err.Error(), resp.StatusCode)
		}
		require.Nil(t, err)
		return ws
	})
	s.After(func(t *testcase.T) { require.Nil(t, ws(t).Close()) })

	subject := func(t *testcase.T) httpapi.EnrollmentResponseBody {
		require.Nil(t, ws(t).WriteJSON(t.I(`request`)))
		var respBody httpapi.EnrollmentResponseBody
		require.Nil(t, ws(t).ReadJSON(&respBody))
		return respBody
	}

	s.When(`request has pilotID`, func(s *testcase.Spec) {
		s.Let(`request`, func(t *testcase.T) interface{} {
			return httpapi.IsFeatureEnabledRequestPayload{Feature: GetFeatureFlagName(t),
				PilotID: GetExternalPilotID(t)}
		})

		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Save(context.Background(), GetFeatureFlag(t)))
			require.Nil(t, GetStorage(t).Save(context.Background(), GetPilot(t)))
		})

		s.Then(`it will reply with the enrollment`, func(t *testcase.T) {
			require.Equal(t, t.I(`PilotEnrollment`).(bool), subject(t).Enrollment)
		})
	})

	s.When(`request has no pilotID`, func(s *testcase.Spec) {
		s.Let(`request`, func(t *testcase.T) interface{} {
			return httpapi.IsFeatureGloballyEnabledRequestPayload{Feature: GetFeatureFlagName(t)}
		})

		s.Let(`rnd`, func(t *testcase.T) interface{} {
			return 99 + rand.Intn(2)
		})

		s.Before(func(t *testcase.T) {
			ff := GetFeatureFlag(t)
			ff.Rollout.Strategy.Percentage = t.I(`rnd`).(int)
			require.Nil(t, GetStorage(t).Save(context.Background(), ff))
		})

		s.Then(`it will reply with the enrollment`, func(t *testcase.T) {
			require.Equal(t, t.I(`rnd`).(int) == 100, subject(t).Enrollment)
		})
	})
}

func withAccessLogging(t *testcase.T, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf(`METHOD=%s PATH=%s`, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
