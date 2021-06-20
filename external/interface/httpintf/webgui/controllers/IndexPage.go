package controllers

import (
	"github.com/toggler-io/toggler/docs"
	"html/template"
	"net/http"

	"github.com/russross/blackfriday"
)

func (ctrl *Controller) IndexPage(w http.ResponseWriter, r *http.Request) {
	markdownBytes, err := docs.FS.ReadFile(`README.md`)
	if err != nil {
		const code = http.StatusNotFound
		http.Error(w, http.StatusText(code), code)
		return
	}
	ctrl.Render(w, `/doc/show.html`, template.HTML(blackfriday.Run(markdownBytes)))
}
