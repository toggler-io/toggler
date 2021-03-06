package swagger

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/toggler-io/toggler/external/interface/httpintf/httputils"
)

func HandleSwaggerConfigJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(200)
	_, _ = w.Write(configJSON)
}

func HandleSwaggerUI() http.Handler {
	var onceFetchSchema sync.Once
	var schema string

	createURL := func(r *http.Request, scheme string) string {
		// NICE_TO_HAVE: remove dependency on "/swagger" base path
		return fmt.Sprintf(`%s://%s/swagger/swagger.json`, scheme, r.Host)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case `/`, `/index.html`:
			t := template.New(`index.html`)
			t, err := t.ParseFS(uiFS, "index.html")
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
			http.FileServer(http.FS(uiFS)).ServeHTTP(w, r)
		}
	})
}
