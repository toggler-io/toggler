package httpgui

import (
	"bytes"
	"context"
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httpapi"
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httpgui/assets"
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httpgui/views"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/security"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
)

//go:generate esc -o ./assets/fs.go -ignore fs.go -pkg assets -prefix assets ./assets
//go:generate esc -o ./views/fs.go  -ignore fs.go -pkg views  -prefix views  ./views

func NewServeMux(uc *usecases.UseCases) *ServeMux {
	mux := &ServeMux{ServeMux: http.NewServeMux(), UseCases: uc,}
	mux.Handle(`/assets/`, http.StripPrefix(`/assets`, assetsFS()))
	mux.HandleFunc(`/`, mux.indexPage)
	mux.HandleFunc(`/update`, mux.updatePage)
	mux.HandleFunc(`/sample`, mux.samplePage)
	return mux
}

func assetsFS() http.Handler {
	return http.FileServer(assets.FS(false))
}

type ServeMux struct {
	*http.ServeMux
	*usecases.UseCases
}

func (mux *ServeMux) render(w http.ResponseWriter, tempName string, data interface{}) {
	tempRawStr := views.FSMustString(false, tempName)
	temp, err := template.New(tempName).Parse(tempRawStr)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf := bytes.NewBuffer([]byte{})

	if err := temp.Execute(buf, data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Println(err)
	}

	return
}

func (mux *ServeMux) showPage(w http.ResponseWriter, r *http.Request) {
	mux.render(w, `/show.html`, rollouts.FeatureFlag{})
}

func (mux *ServeMux) samplePage(w http.ResponseWriter, r *http.Request) {
	mux.render(w, `/x.html`, rollouts.FeatureFlag{})
}

func (mux *ServeMux) updatePage(w http.ResponseWriter, r *http.Request) {
	rr := httptest.NewRecorder()

	token, err := security.NewIssuer(mux.Storage.(usecases.Storage)).CreateNewToken(`test`, nil, nil)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	pu, err := mux.UseCases.ProtectedUsecases(token.Token)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx := context.WithValue(r.Context(), `ProtectedUsecases`, pu)
	r = r.WithContext(ctx)

	httpapi.NewServeMux(mux.UseCases).UpdateFeatureFlagFORM(rr, r)
	http.Redirect(w, r, `/`, http.StatusFound)
}

func (mux *ServeMux) indexPage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case `/`, `/index`:

		var flags = []*rollouts.FeatureFlag{
			{Name: `test-1`},
			{Name: `test-2`},
			{Name: `test-3`},
		}

		fs, err := mux.fakePU().ListFeatureFlags()

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		for _, f := range fs {
			flags = append(flags, f)
		}

		mux.render(w, `/index.html`, flags)

	default:
		http.NotFound(w, r)
	}
}

func (mux *ServeMux) fakePU() *usecases.ProtectedUsecases {

	token, err := security.NewIssuer(mux.Storage.(usecases.Storage)).CreateNewToken(`test`, nil, nil)

	if err != nil {
		panic(err)
	}

	pu, err := mux.UseCases.ProtectedUsecases(token.Token)

	if err != nil {
		panic(err)
	}

	return pu

}
