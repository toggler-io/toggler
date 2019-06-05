package controllers

import (
	"context"
	"fmt"
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httpapi"
	"net/http"
	"net/http/httptest"
)

func (ctrl *Controller) CreatePage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ctrl.Render(w, `/create.html`, nil)

	case http.MethodPost:
		// HACK
		pu := ctrl.GetProtectedUsecases(r)
		ctx := context.WithValue(r.Context(), `ProtectedUsecases`, pu)
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()
		httpapi.NewServeMux(ctrl.UseCases).CreateFeatureFlagFORM(rr, r)

		fmt.Println(rr.Code, rr.Body.String())

		http.Redirect(w, r, `/`, http.StatusFound)

	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}
