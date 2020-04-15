package httputils

import (
	"context"
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

		// Deprecated, remove after no implementation use the protected use-cases object
		pu, err := uc.ProtectedUsecases(r.Context(), token)

		if err == usecases.ErrInvalidToken {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		// Deprecated
		r = r.WithContext(context.WithValue(r.Context(), `ProtectedUseCases`, pu))

		next.ServeHTTP(w, r)

	})
}
