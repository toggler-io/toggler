package httpws

import (
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
)

func NewHandler(uc *toggler.UseCases) http.Handler {
	mux := http.NewServeMux()

	ctrl := Controller{
		UseCases: uc,
		Upgrader: &websocket.Upgrader{},
	}

	mux.Handle(`/`, httputils.AuthMiddleware(http.HandlerFunc(ctrl.WebsocketHandler), uc, http.Error))

	return mux
}
