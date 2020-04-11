package httpws_test

import (
	"context"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/external/interface/httpintf/httpapi"
	"github.com/toggler-io/toggler/external/interface/httpintf/httpws"
	. "github.com/toggler-io/toggler/testing"
)

func TestWebsocket(t *testing.T) {
	s := testcase.NewSpec(t)
	SetUp(s)

	server := func(t *testcase.T) *httptest.Server { return t.I(`server`).(*httptest.Server) }
	s.Let(`server`, func(t *testcase.T) interface{} {
		return httptest.NewServer(httpws.NewHandler(ExampleUseCases(t)))
	})
	s.After(func(t *testcase.T) { server(t).Close() })

	s.Let(`url`, func(t *testcase.T) interface{} {
		url := server(t).URL
		url = "ws" + strings.TrimPrefix(url, "http")
		return url
	})

	s.Let(`TokenString`, func(t *testcase.T) interface{} {
		tSTR, _ := CreateToken(t, `manager`)
		return tSTR
	})

	ws := func(t *testcase.T) *websocket.Conn { return t.I(`ws`).(*websocket.Conn) }
	s.Let(`ws`, func(t *testcase.T) interface{} {
		url := t.I(`url`).(string)

		rHeader := make(http.Header)
		rHeader.Set(`X-App-Token`, t.I(`TokenString`).(string))

		ws, resp, err := websocket.DefaultDialer.Dial(url, rHeader)
		if err == websocket.ErrBadHandshake && resp != nil && err != nil {
			t.Logf(`target url is %q`, url)
			t.Fatalf(`%s with HTTP status code %d`, err.Error(), resp.StatusCode)
		}
		require.Nil(t, err)
		return ws
	})
	s.After(func(t *testcase.T) { require.Nil(t, ws(t).Close()) })

	subject := func(t *testcase.T, resp interface{}) {
		var req httpws.WebsocketRequestPayload
		req.Operation = t.I(`operation`).(string)
		req.Data = t.I(`data`)
		require.Nil(t, ws(t).WriteJSON(req))
		require.Nil(t, ws(t).ReadJSON(resp))
	}

	s.When(`request has pilotID`, func(s *testcase.Spec) {
		s.Let(`operation`, func(t *testcase.T) interface{} {
			return `IsFeatureEnabled`
		})

		s.Let(`data`, func(t *testcase.T) interface{} {
			return httpws.IsFeatureEnabledRequestPayload{
				Feature: ExampleReleaseFlagName(t),
				PilotID: GetExternalPilotID(t),
			}
		})

		s.Before(func(t *testcase.T) {
			require.Nil(t, ExampleStorage(t).Create(context.Background(), ExampleReleaseFlag(t)))
			require.Nil(t, ExampleStorage(t).Create(context.Background(), ExamplePilot(t)))
		})

		s.Then(`it will reply with the enrollment`, func(t *testcase.T) {
			var resp httpapi.EnrollmentResponseBody
			subject(t, &resp)
			require.Equal(t, t.I(`PilotEnrollment`).(bool), resp.Enrollment)
		})
	})

	s.When(`request has no pilotID`, func(s *testcase.Spec) {
		s.Let(`operation`, func(t *testcase.T) interface{} {
			return `IsFeatureGloballyEnabled`
		})

		s.Let(`data`, func(t *testcase.T) interface{} {
			return httpapi.IsFeatureGloballyEnabledRequestBody{Feature: ExampleReleaseFlagName(t)}
		})

		s.Let(`rnd`, func(t *testcase.T) interface{} {
			return 99 + rand.New(rand.NewSource(time.Now().Unix())).Intn(2)
		})

		s.Before(func(t *testcase.T) {
			ff := ExampleReleaseFlag(t)
			ff.Rollout.Strategy.Percentage = t.I(`rnd`).(int)
			require.Nil(t, ExampleStorage(t).Create(context.Background(), ff))
		})

		s.Then(`it will reply with the enrollment`, func(t *testcase.T) {
			var resp httpapi.EnrollmentResponseBody
			subject(t, &resp)
			expectedEnrollment := t.I(`rnd`).(int) == 100
			require.Equal(t, expectedEnrollment, resp.Enrollment)
		})
	})
}
