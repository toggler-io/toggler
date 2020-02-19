package httppath_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/toggler-io/toggler/extintf/httpintf/httppath"
)

type ctxKeyResourceID struct{}

type ResourceHandler struct {
	ServeMux *http.ServeMux
}

func NewResourceHandler() *ResourceHandler {
	h := &ResourceHandler{ServeMux: http.NewServeMux()}
	h.ServeMux.HandleFunc(`/edit`, h.edit)
	return h
}

// http path for this could be /resources/:resource_id/edit
// but anything is fine as long the context has the resourceID
func (h *ResourceHandler) edit(w http.ResponseWriter, r *http.Request) {
	resourceID := r.Context().Value(ctxKeyResourceID{}).(string)
	_, _ = fmt.Fprintf(w, `resource id is: %s`, resourceID)
}

func (h *ResourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ServeMux.ServeHTTP(w, r)
}

func ExampleUnshiftPathParam() {
	var withResourceID = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r, param := httppath.UnshiftPathParam(r)
			// verify resource id is valid and present
			// verify authority of the requester to this resourceID
			r = r.WithContext(context.WithValue(r.Context(), ctxKeyResourceID{}, param)) // add request context
			next.ServeHTTP(w, r)
		})
	}

	mux := http.NewServeMux()
	mux.Handle(`/resources/`, http.StripPrefix(`/resources`, withResourceID(NewResourceHandler())))
}
