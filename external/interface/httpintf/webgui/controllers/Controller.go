package controllers

import (
	"log"
	"net/http"

	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/interface/httpintf/webgui/views"
)

func NewController(uc *toggler.UseCases) (*Controller, error) {
	renderer, err := NewHttpFileSystemRenderer(views.FS(false))
	if err != nil {
		return nil, err
	}

	return &Controller{
		UseCases: uc,
		Renderer: renderer,
	}, nil
}

type Controller struct {
	*toggler.UseCases
	Renderer Renderer
}

func (c *Controller) Render(w http.ResponseWriter, tmpl string, data interface{}) {
	c.Renderer.Render(w, tmpl, data)
}

type Renderer interface {
	Render(http.ResponseWriter, string, interface{})
}

func (ctrl *Controller) handleError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}

	log.Println(err.Error())
	http.Redirect(w, r, r.URL.Path, http.StatusFound)
	return true
}
