package httpgui

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"net/http"
)

func (mux *ServeMux) IndexPage(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case `/`, `/index`:

		var flags = []*rollouts.FeatureFlag{
			{Name: `test-1`},
			{Name: `test-2`},
			{Name: `test-3`},
		}

		fs, err := mux.fakePU().ListFeatureFlags()

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		for _, f := range fs {
			flags = append(flags, f)
		}

		mux.render(w, `/index.html`, flags)

	default:
		http.NotFound(w, r)
	}
}
