package httputils

import (
	"net/http"

	"github.com/toggler-io/toggler/usecases"
)

type ErrorWriterFunc func(w http.ResponseWriter, error string, code int)

func AuthMiddleware(next http.Handler, uc *usecases.UseCases, errorWriterFunc ErrorWriterFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := GetAppToken(r)

		if err != nil {
			errorWriterFunc(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		valid, err := uc.Doorkeeper.VerifyTextToken(r.Context(), token)

		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		if !valid {
			code := http.StatusUnauthorized
			http.Error(w, http.StatusText(code), code)
			return
		}

		next.ServeHTTP(w, r)

	})
}
