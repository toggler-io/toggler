package httpintf

import (
	"net/http"

	"github.com/adamluzsi/toggler/extintf/httpintf/httpapi"
	"github.com/adamluzsi/toggler/extintf/httpintf/webgui"
	"github.com/adamluzsi/toggler/usecases"
)

func NewServeMux(uc *usecases.UseCases) *ServeMux {
	mux := http.NewServeMux()

	mux.Handle(`/api/`, http.StripPrefix(`/api`, httpapi.NewServeMux(uc)))
	mux.Handle(`/`, webgui.NewServeMux(uc))

	return &ServeMux{
		ServeMux: mux,
		UseCases: uc,
	}
}

type ServeMux struct {
	*http.ServeMux
	*usecases.UseCases
}
