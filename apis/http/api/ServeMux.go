package api

import (
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
)

func NewServeMux(uc *usecases.UseCases) *ServeMux {
	mux := &ServeMux{
		ServeMux: http.NewServeMux(),
		UseCases: uc,
	}

	mux.Handle(`/is-feature-enabled-for`, http.HandlerFunc(mux.IsFeatureEnabledFor))

	return mux
}

type ServeMux struct {
	*http.ServeMux
	*usecases.UseCases
}
