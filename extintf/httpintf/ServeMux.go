package httpintf

import (
	"github.com/adamluzsi/toggler/extintf/httpintf/swagger"
	"net/http"

	"github.com/adamluzsi/toggler/extintf/httpintf/httpapi"
	"github.com/adamluzsi/toggler/extintf/httpintf/webgui"
	"github.com/adamluzsi/toggler/usecases"
)

func NewServeMux(uc *usecases.UseCases) *ServeMux {
	mux := http.NewServeMux()

	mux.Handle(`/api/v1/`, letsCORSit(http.StripPrefix(`/api/v1`, httpapi.NewServeMux(uc))))
	mux.Handle(`/`, webgui.NewServeMux(uc))
	mux.Handle(`/swagger.json`, letsCORSit(http.HandlerFunc(swagger.HandleSwaggerConfigJSON)))
	mux.Handle(`/swagger-ui/`, http.StripPrefix(`/swagger-ui`, http.HandlerFunc(swagger.HandleSwaggerUI)))

	return &ServeMux{
		ServeMux: mux,
		UseCases: uc,
	}
}

type ServeMux struct {
	*http.ServeMux
	*usecases.UseCases
}

func letsCORSit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(`Access-Control-Request-Method`, `*`)
		w.Header().Set(`Access-Control-Allow-Headers`, `*`)
		w.Header().Set(`Access-Control-Allow-Origin`, `*`)
		if r.Method == http.MethodOptions {
			w.WriteHeader(200)
			return
		}

		next.ServeHTTP(w, r)
	})
}
