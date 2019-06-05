package controllers

import (
	"context"
	"fmt"
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httpapi"
	"net/http"
	"net/http/httptest"
)

// super hacky implementation for the sake of POC
func (ctrl *Controller) UpdatePage(w http.ResponseWriter, r *http.Request) {
	pu := ctrl.GetProtectedUsecases(r)
	ctx := context.WithValue(r.Context(), `ProtectedUsecases`, pu)
	r = r.WithContext(ctx)
	rr := httptest.NewRecorder()
	httpapi.NewServeMux(ctrl.UseCases).UpdateFeatureFlagFORM(rr, r)
	fmt.Println(rr.Code, rr.Body.String())

	http.Redirect(w, r, `/`, http.StatusFound)
}
