package swagger

import (
	"fmt"
	"github.com/adamluzsi/toggler/extintf/httpintf/httputils"
	"github.com/adamluzsi/toggler/extintf/httpintf/swagger/specfs"
	"github.com/adamluzsi/toggler/extintf/httpintf/swagger/uifs"
	"html/template"
	"net/http"
	"sync"
)

func HandleSwaggerConfigJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(200)
	_, _ = w.Write(specfs.FSMustByte(false, `/swagger.json`))
}

func HandleSwaggerUI() http.Handler {
	var onceFetchSchema sync.Once
	var schema string

	createURL := func(r *http.Request, scheme string) string {
		return fmt.Sprintf(`%s://%s/swagger.json`, scheme, r.Host)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		switch r.URL.Path {
		case `/`, `/index.html`:
			t, err := template.New(`swagger-ui`).Parse(uifs.FSMustString(false, `/index.html`))
			if httputils.HandleError(w, err, http.StatusInternalServerError) {
				return
			}

			onceFetchSchema.Do(func() {

				u := createURL(r, `https`)

				_, err := http.Get(u)

				if err == nil {
					schema = `https`
				} else {
					schema = `http`
				}

			})
			
			data := struct{ ConfigURL string }{}
			data.ConfigURL = createURL(r, schema)

			if httputils.HandleError(w, t.Execute(w, data), http.StatusInternalServerError) {
				return
			}

			return

		default:
			http.FileServer(uifs.FS(false)).ServeHTTP(w, r)
		}
	})
}
