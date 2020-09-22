package webgui

import (
	"net/http"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/interface/httpintf/webgui/cookies"
)

func NewHandler(uc *toggler.UseCases) (http.Handler, error) {
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
