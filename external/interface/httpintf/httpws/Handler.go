package httpws

import (
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
	"github.com/toggler-io/toggler/usecases"
)

func NewHandler(uc *usecases.UseCases) http.Handler {
	mux := http.NewServeMux()

	ctrl := Controller{
		UseCases: uc,
		Upgrader: &websocket.Upgrader{},
	}

	mux.Handle(`/`, httputils.AuthMiddleware(http.HandlerFunc(ctrl.WebsocketHandler), uc))

	return mux
}
