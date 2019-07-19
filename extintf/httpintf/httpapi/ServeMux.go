package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/toggler/extintf/httpintf/httputils"
	"github.com/adamluzsi/toggler/usecases"
)

func NewServeMux(uc *usecases.UseCases) *ServeMux {
	mux := &ServeMux{ServeMux: http.NewServeMux(), UseCases: uc}

	featureAPI := buildFeatureAPI(mux)
	mux.Handle(`/rollout/`, http.StripPrefix(`/rollout`, featureAPI))

	return mux
}

func buildFeatureAPI(handlers *ServeMux) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle(`/config.json`, http.HandlerFunc(handlers.RolloutConfigJSON))
	mux.Handle(`/is-enabled.json`, http.HandlerFunc(handlers.IsFeatureEnabledFor))
	mux.Handle(`/is-globally-enabled.json`, http.HandlerFunc(handlers.IsFeatureGloballyEnabled))
	mux.Handle(`/flag/`, http.StripPrefix(`/flag`, buildFlagAPI(handlers)))
	return mux
}

func buildFlagAPI(handlers *ServeMux) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(`/create.form`, http.HandlerFunc(handlers.CreateFeatureFlagFORM))
	mux.Handle(`/create.json`, http.HandlerFunc(handlers.CreateFeatureFlagJSON))
	mux.Handle(`/update.form`, http.HandlerFunc(handlers.UpdateFeatureFlagFORM))
	mux.Handle(`/update.json`, http.HandlerFunc(handlers.UpdateFeatureFlagJSON))
	mux.Handle(`/list.json`, http.HandlerFunc(handlers.ListFeatureFlags))
	mux.Handle(`/set-enrollment-manually.json`, http.HandlerFunc(handlers.SetPilotEnrollmentForFeature))
	return authMiddleware(handlers.UseCases, mux)
}

type ServeMux struct {
	*http.ServeMux
	*usecases.UseCases
}

func serveJSON(w http.ResponseWriter, data interface{}) {
	buf := bytes.NewBuffer([]byte{})

	if err := json.NewEncoder(buf).Encode(data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(`Content-Type`, `application/json`)
	w.WriteHeader(200)

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Println(err)
	}
}

func handleError(w http.ResponseWriter, err error, errCode int) (errorWasHandled bool) {
	if err == usecases.ErrInvalidToken {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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
