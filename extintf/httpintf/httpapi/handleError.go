package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/adamluzsi/frameless"

	"github.com/toggler-io/toggler/usecases"
)

// ErrorResponse will contains a response about request that had some kind of problem.
// The details will be included in the body.
// swagger:response errorResponse
type ErrorResponse struct {
	// Body describe and error that meant to be consumed by a software engineer.
	// in: body
	Body struct {
		Error Error `json:"error"`
	}
}

// Error contains the details of the error
type Error struct {
	// The constant code of the error that can be used for localisation
	// Example: 401
	Code int `json:"code"`
	// The message that describe the error to the developer who do the integration.
	// Not meant to be propagated to the end-user.
	// The Message may change in the future, it it helps readability,
	// please do not rely on the content in any way other than just reading it.
	Message string `json:"message"`
}

func handleError(w http.ResponseWriter, err error, errCode int) (errorWasHandled bool) {

	toErrResp := func(code int) []byte {
		var errResp ErrorResponse
		errResp.Body.Error.Code = code

		if 400 <= errCode && errCode < 500 {
			errResp.Body.Error.Message = err.Error()
		} else {
			errResp.Body.Error.Message = http.StatusText(code)
		}

		body, mErr := json.Marshal(errResp.Body)
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
		w.WriteHeader(errCode)
		_, _ = w.Write(toErrResp(http.StatusUnauthorized))
		return true
	}

	return false

}
