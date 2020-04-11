package cookies

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/toggler-io/toggler/usecases"
)

const authTokenCookieName = `auth-token`

func LookupAuthToken(r *http.Request) ([]byte, bool, error) {
	cookie, err := r.Cookie(authTokenCookieName)

	if err == http.ErrNoCookie {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if cookie == nil {
		return nil, false, nil
	}

	return []byte(cookie.Value), true, nil
}

func SetAuthToken(w http.ResponseWriter, token []byte) {
	http.SetCookie(w, &http.Cookie{
		Name:    authTokenCookieName,
		Value:   string(token),
		Expires: time.Now().AddDate(1, 0, 0),
	})
}

type AuthTokenContextKey struct{}

func WithAuthTokenMiddleware(next http.Handler, uc *usecases.UseCases, signInURL string, exceptions []string) http.Handler {
	mux := http.NewServeMux()

	// register exceptions
	for _, exception := range append([]string{signInURL}, exceptions...) {
		mux.Handle(exception, next)
	}

	mux.Handle(`/`, &AuthTokenMiddleware{
		Next:       next,
		RedirectTo: signInURL,
		Doorkeeper: uc.Doorkeeper,
	})

	return mux
}

type AuthTokenMiddleware struct {
	Next       http.Handler
	RedirectTo string
	Doorkeeper interface {
		VerifyTextToken(ctx context.Context, textToken string) (bool, error)
	}
}

func (mw *AuthTokenMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(`DEBUG`, r.URL.Path)
	token, ok, err := LookupAuthToken(r)
	if err != nil {
		log.Println(`ERROR`, err.Error())
		const code = http.StatusInternalServerError
		http.Error(w, http.StatusText(code), code)
		return
	}
	if !ok {
		http.Redirect(w, r, mw.RedirectTo, http.StatusFound)
		return
	}

	valid, err := mw.Doorkeeper.VerifyTextToken(r.Context(), string(token))
	if err != nil {
		log.Println(`ERROR`, err.Error())
		const code = http.StatusInternalServerError
		http.Error(w, http.StatusText(code), code)
		return
	}
	if !valid {
		http.Redirect(w, r, mw.RedirectTo, http.StatusFound)
		return
	}

	mw.Next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), AuthTokenContextKey{}, token)))
}
