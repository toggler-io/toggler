package controllers

import (
	"bytes"
	"github.com/adamluzsi/toggler/extintf/httpintf/webgui/views"
	"github.com/adamluzsi/toggler/usecases"
	"html/template"
	"log"
	"net/http"
)

func NewController(uc *usecases.UseCases) *Controller {
	return &Controller{UseCases: uc, Render: renderFunc,}
}

type Controller struct {
	*usecases.UseCases
	Render func(http.ResponseWriter, string, interface{})
}

//TODO: cache templates with closure
func renderFunc(w http.ResponseWriter, tempName string, data interface{}) {
	layoutRawStr := views.FSMustString(false, `/layout.html`)
	pageRawStr := views.FSMustString(false, tempName)

	var err error
	tmpl := template.New(``)

	if tmpl, err = tmpl.New(`page`).Parse(layoutRawStr); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if tmpl, err = tmpl.New(`content`).Parse(pageRawStr); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf := bytes.NewBuffer([]byte{})

	if err := tmpl.ExecuteTemplate(buf, `page`, data); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Println(err)
	}

	return
}

func (ctrl *Controller) GetProtectedUsecases(r *http.Request) *usecases.ProtectedUsecases {
	return r.Context().Value(`*usecases.ProtectedUsecases`).(*usecases.ProtectedUsecases)
}

func (ctrl *Controller) handleError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}

	log.Println(err.Error())
	http.Redirect(w,r, r.URL.Path, http.StatusFound)
	return true
}
