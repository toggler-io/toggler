package httpintf

import (
	"github.com/toggler-io/toggler/extintf/httpintf/httputils"
	"github.com/toggler-io/toggler/extintf/httpintf/swagger"
	"net/http"

	"github.com/toggler-io/toggler/extintf/httpintf/httpapi"
	"github.com/toggler-io/toggler/extintf/httpintf/webgui"
	"github.com/toggler-io/toggler/usecases"
)

func NewServeMux(uc *usecases.UseCases) (*ServeMux, error) {
	mux := http.NewServeMux()

	mux.Handle(`/api/`, httputils.CORS(http.StripPrefix(`/api`, httpapi.NewHandler(uc))))

	ui, err := webgui.NewHandler(uc)
	if err != nil {
		return nil, err
	}

	mux.Handle(`/`, ui)
	mux.Handle(`/swagger.json`, httputils.CORS(http.HandlerFunc(swagger.HandleSwaggerConfigJSON)))
	mux.Handle(`/swagger-ui/`, http.StripPrefix(`/swagger-ui`, swagger.HandleSwaggerUI()))

	return &ServeMux{
		ServeMux: mux,
		UseCases: uc,
	}, nil
}

type ServeMux struct {
	*http.ServeMux
	*usecases.UseCases
}
