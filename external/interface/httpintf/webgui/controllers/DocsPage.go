package controllers

import (
	"html/template"
	"net/http"

	"github.com/russross/blackfriday"

	"github.com/toggler-io/toggler/external/interface/httpintf/webgui/controllers/docspages"
)

//go:generate esc -o ./docspages/fs.go -pkg docspages -prefix "${WDP}" "${WDP}/docs"

func (ctrl *Controller) DocsPage(w http.ResponseWriter, r *http.Request) {
	markdownBytes := docspages.FSMustByte(false, r.URL.Path)
	pageContent := template.HTML(blackfriday.Run(markdownBytes))
	ctrl.Render(w, `/doc/show.html`, pageContent)
}

func (ctrl *Controller) DocsAssets(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write(docspages.FSMustByte(false, r.URL.Path))
}
