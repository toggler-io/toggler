package controllers

import (
	"github.com/toggler-io/toggler/docs"
	"html/template"
	"net/http"
	"strings"

	"github.com/russross/blackfriday"
)

func (ctrl *Controller) DocsPage(w http.ResponseWriter, r *http.Request) {
	markdownBytes, err := docs.FS.ReadFile(ctrl.docPath(r))
	if err != nil {
		const code = http.StatusNotFound
		http.Error(w, http.StatusText(code), code)
		return
	}
	pageContent := template.HTML(blackfriday.Run(markdownBytes))
	ctrl.Render(w, `/doc/show.html`, pageContent)
}

func (ctrl *Controller) DocsAssets(w http.ResponseWriter, r *http.Request) {
	bs, err := docs.FS.ReadFile(ctrl.docPath(r))
	if err != nil {
		const code = http.StatusNotFound
		http.Error(w, http.StatusText(code), code)
		return
	}
	_, _ = w.Write(bs)
}

func (ctrl *Controller) docPath(r *http.Request) string {
	return strings.TrimPrefix(r.URL.Path, "/docs/")
}
