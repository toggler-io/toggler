package swagger

import (
	"net/http"

	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle(`/api.json`, httputils.CORS(http.HandlerFunc(HandleSwaggerConfigJSON)))
	mux.Handle(`/ui/`, http.StripPrefix(`/ui`, HandleSwaggerUI()))
	return mux
}
