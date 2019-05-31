package httpintf

import (
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httpapi"
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httpgui"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
)

func NewServeMux(uc *usecases.UseCases) *ServeMux {
	mux := http.NewServeMux()

	mux.Handle(`/api/`, http.StripPrefix(`/api`, httpapi.NewServeMux(uc)))
	mux.Handle(`/`, httpgui.NewServeMux(uc))

	return &ServeMux{
		ServeMux: mux,
		UseCases: uc,
	}
}

type ServeMux struct {
	*http.ServeMux
	*usecases.UseCases
}
