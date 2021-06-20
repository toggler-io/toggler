package httputils

import (
	"log"
	"net/http"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/interface/httpintf/webgui/cookies"
)

func GetAppToken(r *http.Request) (string, error) {
	var token string
	token = r.Header.Get(`X-App-Token`)

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
	if err != nil {
		log.Println("ERROR", err.Error())
	}
	if err == toggler.ErrInvalidToken {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return true
	}
	if err != nil {
		http.Error(w, http.StatusText(errCode), errCode)
		return true
	}
	return false
}
