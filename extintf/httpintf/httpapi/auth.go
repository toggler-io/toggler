package httpapi

import (
	"context"
	"net/http"

	"github.com/toggler-io/toggler/extintf/httpintf/httputils"
	"github.com/toggler-io/toggler/usecases"
)

func authMiddleware(next http.Handler, uc *usecases.UseCases) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := httputils.GetAuthToken(r)

		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		valid, err := uc.Doorkeeper.VerifyTextToken(r.Context(), token)

		if handleError(w, err, http.StatusInternalServerError) {
			return
		}

		if !valid {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Deprecated, remove after no implementation use the protected use-cases object
		pu, err := uc.ProtectedUsecases(r.Context(), token)

		if err == usecases.ErrInvalidToken {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if handleError(w, err, http.StatusInternalServerError) {
			return
		}

		// Deprecated
		r = r.WithContext(context.WithValue(r.Context(), `ProtectedUseCases`, pu))

		next.ServeHTTP(w, r)

	})
}