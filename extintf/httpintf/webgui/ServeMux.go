package webgui

import (
	"context"
	"net/http"

	"github.com/adamluzsi/toggler/extintf/httpintf/httputils"
	"github.com/adamluzsi/toggler/extintf/httpintf/webgui/assets"
	"github.com/adamluzsi/toggler/extintf/httpintf/webgui/controllers"
	"github.com/adamluzsi/toggler/usecases"
)

//go:generate esc -o ./assets/fs.go -ignore fs.go -pkg assets -prefix assets ./assets
//go:generate esc -o ./views/fs.go  -ignore fs.go -pkg views  -prefix views  ./views

func NewServeMux(uc *usecases.UseCases) *ServeMux {
	ctrl := controllers.NewController(uc)
	mux := &ServeMux{ServeMux: http.NewServeMux(), UseCases: uc}
	mux.Handle(`/assets/`, http.StripPrefix(`/assets`, assetsFS()))
	mux.Handle(`/`, authorized(uc, ctrl.IndexPage))
	mux.Handle(`/flag`, authorized(uc, ctrl.FlagPage))
	mux.Handle(`/flag/`, authorized(uc, ctrl.FlagPage))
	mux.Handle(`/docs/`, authorized(uc, ctrl.DocsPage))
	mux.Handle(`/pilot/`, authorized(uc, ctrl.PilotPage))
	mux.HandleFunc(`/login`, ctrl.LoginPage)
	return mux
}

type ServeMux struct {
	*http.ServeMux
	*usecases.UseCases
}

func assetsFS() http.Handler {
	return http.FileServer(assets.FS(false))
}

func authorized(uc *usecases.UseCases, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := httputils.GetAuthToken(r)

		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		pu, err := uc.ProtectedUsecases(context.TODO(), token)

		if err == usecases.ErrInvalidToken {
			http.Redirect(w, r, `/login`, http.StatusFound)
			return
		}

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), `*usecases.ProtectedUsecases`, pu))

		next.ServeHTTP(w, r)

	})
}
