package swagger

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/toggler-io/toggler/extintf/httpintf/httputils"
	"github.com/toggler-io/toggler/extintf/httpintf/swagger/specfs"
	"github.com/toggler-io/toggler/extintf/httpintf/swagger/uifs"
)

func HandleSwaggerConfigJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(200)
	_, _ = w.Write(specfs.FSMustByte(false, `/api.json`))
}

func HandleSwaggerUI() http.Handler {
	var onceFetchSchema sync.Once
	var schema string

	createURL := func(r *http.Request, scheme string) string {
		// NICE_TO_HAVE: remove dependency on "/swagger" base path
		return fmt.Sprintf(`%s://%s/swagger/api.json`, scheme, r.Host)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
