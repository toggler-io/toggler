package httputils

import (
	"github.com/adamluzsi/FeatureFlags/usecases"
	"net/http"
)

func GetAuthToken(r *http.Request) (string, error) {
	token := r.URL.Query().Get(`token`)

	if token == `` {
		token = r.Header.Get(`X-Auth-Token`)
	}

	if token == `` {
		cookie, err := r.Cookie(`token`)

		if err != http.ErrNoCookie && err != nil {
			return "", err
		}

		if cookie != nil {
			token = cookie.Value
		}
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
