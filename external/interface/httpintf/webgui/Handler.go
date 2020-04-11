package webgui

import (
	"net/http"

	"github.com/toggler-io/toggler/external/interface/httpintf/webgui/cookies"
	"github.com/toggler-io/toggler/usecases"
)

func NewHandler(uc *usecases.UseCases) (http.Handler, error) {
	mux, err := NewServeMux(uc)
	if err != nil {
		return nil, err
	}

	return cookies.WithAuthTokenMiddleware(mux, uc, `/login`, []string{
		`/swagger-ui/`,
		`/assets/`,
		`/favicon.ico`,
	}), nil
}
