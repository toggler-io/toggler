package httpintf

import (
	"net/http"

	"github.com/adamluzsi/gorest"

	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
	"github.com/toggler-io/toggler/external/interface/httpintf/swagger"

	"github.com/toggler-io/toggler/external/interface/httpintf/httpapi"
	"github.com/toggler-io/toggler/external/interface/httpintf/webgui"
	"github.com/toggler-io/toggler/usecases"
)

func NewServeMux(uc *usecases.UseCases) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	mux.Handle(`/api/`, httputils.CORS(http.StripPrefix(`/api`, httpapi.NewHandler(uc))))
	//mux.Handle(`/ws/`, httputils.CORS(http.StripPrefix(`/ws`, httpws.NewHandler(uc))))

	ui, err := webgui.NewHandler(uc)
	if err != nil {
		return nil, err
	}

	mux.Handle(`/`, ui)

	// TODO: fix the behavior where "/swagger/ui" redirects to "/ui"
	gorest.Mount(mux, `/swagger`, swagger.NewHandler())

	return mux, nil
}
