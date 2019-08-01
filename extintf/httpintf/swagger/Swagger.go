package swagger

import (
	"github.com/adamluzsi/toggler/extintf/httpintf/httputils"
	"github.com/adamluzsi/toggler/extintf/httpintf/swagger/specfs"
	"github.com/adamluzsi/toggler/extintf/httpintf/swagger/uifs"
	"html/template"
	"net/http"
)

func HandleSwaggerConfigJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(200)
	_, _ = w.Write(specfs.FSMustByte(false, `/swagger.json`))
}

func HandleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case `/`, `/index.html`:
		t, err := template.New(`swagger-ui`).Parse(uifs.FSMustString(false, `/index.html`))
		if httputils.HandleError(w, err, http.StatusInternalServerError) {
			return
		}

		data := struct{ ConfigURL string }{}
		data.ConfigURL = `/swagger.json`

		if httputils.HandleError(w, t.Execute(w, data), http.StatusInternalServerError) {
			return
		}

		return

	default:
		http.FileServer(uifs.FS(false)).ServeHTTP(w, r)
	}
}
