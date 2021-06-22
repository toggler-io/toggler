package httputils

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(`Access-Control-Allow-Methods`, `*`)
		w.Header().Set(`Access-Control-Allow-Headers`, `*`)
		w.Header().Set(`Access-Control-Allow-Origin`, `*`)
		if r.Method == http.MethodOptions {
			w.WriteHeader(200)
			return
		}

		next.ServeHTTP(w, r)
	})
}
