package controllers

import (
	"github.com/toggler-io/toggler/extintf/httpintf/webgui/controllers/docspages"
	"github.com/russross/blackfriday"
	"html/template"
	"net/http"
)

func (ctrl *Controller) IndexPage(w http.ResponseWriter, r *http.Request) {
	markdownBytes := docspages.FSMustByte(false, `/docs/README.md`)
	pageContent := template.HTML(blackfriday.Run(markdownBytes))
	ctrl.Render(w, `/doc/show.html`, pageContent)
}
