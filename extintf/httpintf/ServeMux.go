package httpintf

import (
	"net/http"

	"github.com/adamluzsi/gorest"

	"github.com/toggler-io/toggler/extintf/httpintf/httputils"
	"github.com/toggler-io/toggler/extintf/httpintf/httpws"
	"github.com/toggler-io/toggler/extintf/httpintf/swagger"

	"github.com/toggler-io/toggler/extintf/httpintf/httpapi"
	"github.com/toggler-io/toggler/extintf/httpintf/webgui"
	"github.com/toggler-io/toggler/usecases"
)

func NewServeMux(uc *usecases.UseCases) (*ServeMux, error) {
	mux := http.NewServeMux()

	mux.Handle(`/api/`, httputils.CORS(http.StripPrefix(`/api`, httpapi.NewHandler(uc))))
	mux.Handle(`/ws/`, httputils.CORS(http.StripPrefix(`/ws`, httpws.NewHandler(uc))))

	ui, err := webgui.NewHandler(uc)
	if err != nil {
		return nil, err
	}

	mux.Handle(`/`, ui)
	gorest.Mount(mux, `/swagger`, swagger.NewHandler())

	return &ServeMux{
		ServeMux: mux,
		UseCases: uc,
	}, nil
}

type ServeMux struct {
	*http.ServeMux
	*usecases.UseCases
}
