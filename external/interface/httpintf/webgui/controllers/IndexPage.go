package controllers

import (
	"html/template"
	"net/http"

	"github.com/russross/blackfriday"

	"github.com/toggler-io/toggler/external/interface/httpintf/webgui/controllers/docspages"
)

func (ctrl *Controller) IndexPage(w http.ResponseWriter, r *http.Request) {
	markdownBytes := docspages.FSMustByte(false, `/docs/README.md`)
	pageContent := template.HTML(blackfriday.Run(markdownBytes))
	ctrl.Render(w, `/doc/show.html`, pageContent)
}
