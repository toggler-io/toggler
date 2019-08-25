package httpapi

import (
	"context"
	"encoding/json"
	"github.com/adamluzsi/frameless"
	"github.com/toggler-io/toggler/extintf/httpintf/httputils"
	"github.com/toggler-io/toggler/usecases"
	"net/http"
)

// ErrorResponse will contains a response about request that had some kind of problem.
// The details will be included in the body.
// swagger:response errorResponse
type ErrorResponse struct {
	// Error contains the details of the error
	// in: body
	Body ErrorResponseBody
}

// ErrorResponseBody describe and error that meant to be consumed by a software engineer.
type ErrorResponseBody struct {
	// Error contains the details of the error
	Error struct {
		// The constant code of the error that can be used for localisation
		// Example: 401
		Code int `json:"code"`
		// The message that describe the error to the developer who do the integration.
		// Not meant to be propagated to the end-user.
		// The Message may change in the future, it it helps readability,
		// please do not rely on the content in any way other than just reading it.
		Message string `json:"message"`
	} `json:"error"`
}

func handleError(w http.ResponseWriter, err error, errCode int) (errorWasHandled bool) {

	toErrResp := func(code int) []byte {
		var errResp ErrorResponseBody
		errResp.Error.Code = code
		errResp.Error.Message = http.StatusText(code)
		body, mErr := json.Marshal(errResp)
		if mErr != nil {
			panic(mErr)
		}
		return body
	}

	if err == usecases.ErrInvalidToken {

		_, _ = w.Write(toErrResp(http.StatusUnauthorized))
		return true
	}

	if err == frameless.ErrNotFound {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return true
	}

	if err != nil {
		http.Error(w, http.StatusText(errCode), errCode)
		return true
	}

	return false
}

func authMiddleware(uc *usecases.UseCases, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := httputils.GetAuthToken(r)

		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		pu, err := uc.ProtectedUsecases(context.TODO(), token)

		if err == usecases.ErrInvalidToken {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if handleError(w, err, http.StatusInternalServerError) {
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), `ProtectedUsecases`, pu)))

	})
}

