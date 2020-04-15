package swagger

import (
	"net/http"

	"github.com/adamluzsi/gorest"

	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle(`/api.json`, httputils.CORS(http.HandlerFunc(HandleSwaggerConfigJSON)))
	gorest.Mount(mux, `/ui`, HandleSwaggerUI())
	return mux
}
