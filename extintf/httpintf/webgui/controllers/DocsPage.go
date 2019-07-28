package controllers

import (
	"github.com/adamluzsi/toggler/extintf/httpintf/webgui/controllers/docspages"
	"github.com/russross/blackfriday"
	"html/template"
	"net/http"
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
