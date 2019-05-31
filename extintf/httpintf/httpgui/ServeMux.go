package httpgui

import (
	"bytes"
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httpgui/assets"
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf/httpgui/views"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"html/template"
	"log"
	"net/http"
)

//go:generate esc -o ./assets/fs.go -ignore fs.go -pkg assets -prefix assets ./assets
//go:generate esc -o ./views/fs.go  -ignore fs.go -pkg views  -prefix views  ./views

func NewServeMux(uc *usecases.UseCases) *ServeMux {
	mux := &ServeMux{ServeMux: http.NewServeMux(), UseCases: uc,}
	mux.Handle(`/assets/`, http.StripPrefix(`/assets`, assetsFS()))
	mux.HandleFunc(`/`, mux.indexPage)
	mux.HandleFunc(`/show`, mux.showPage)
	mux.HandleFunc(`/update`, mux.updatePage)
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

func (mux *ServeMux) updatePage(w http.ResponseWriter, r *http.Request) {
	mux.render(w, `/show.html`, rollouts.FeatureFlag{})
}

func (mux *ServeMux) indexPage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case `/`, `/index`:
		mux.render(w, `/index.html`, nil)

	default:
		http.NotFound(w, r)
	}
}
