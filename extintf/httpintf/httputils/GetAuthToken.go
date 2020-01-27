package httputils

import (
	"net/http"

	"github.com/toggler-io/toggler/extintf/httpintf/webgui/cookies"
	"github.com/toggler-io/toggler/usecases"
)

func GetAuthToken(r *http.Request) (string, error) {
	var token string
	token = r.Header.Get(`X-Auth-Token`)

	if token == `` {
		token = r.URL.Query().Get(`token`)
	}

	if token == `` {
		tokenBS, ok, err := cookies.LookupAuthToken(r)
		if err != nil {
			return ``, err
		}
		if !ok {
			return ``, nil
		}
		token = string(tokenBS)
	}

	return token, nil
}

func HandleError(w http.ResponseWriter, err error, errCode int) (errorWasHandled bool) {
	if err == usecases.ErrInvalidToken {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return true
	}

	if err != nil {
		http.Error(w, http.StatusText(errCode), errCode)
		return true
	}

	return false
}
