package controllers

import (
	"bytes"
	"github.com/toggler-io/toggler/extintf/httpintf/webgui/views"
	"github.com/toggler-io/toggler/usecases"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

func NewController(uc *usecases.UseCases) *Controller {
	return &Controller{
		UseCases: uc,
		Render:   CreateRenderFunc(views.FS(false)),
	}
}

type Controller struct {
	*usecases.UseCases
	Render func(http.ResponseWriter, string, interface{})
}

//TODO: cache templates with closure
func CreateRenderFunc(fs http.FileSystem) func(w http.ResponseWriter, tempName string, data interface{}) {
	fsString := func(name string) (string, error) {
		f, err := fs.Open(name)
		if err != nil {
			return ``, err
		}
		bs, err := ioutil.ReadAll(f)
		if err != nil {
			return ``, err
		}
		return string(bs), nil
	}

	return func(w http.ResponseWriter, tempName string, data interface{}) {
		var err error

		layoutRawStr, err := fsString(`/layout.html`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pageRawStr, err := fsString(tempName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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
}

func (ctrl *Controller) GetProtectedUsecases(r *http.Request) *usecases.ProtectedUsecases {
	return r.Context().Value(`*usecases.ProtectedUsecases`).(*usecases.ProtectedUsecases)
}

func (ctrl *Controller) handleError(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}

	log.Println(err.Error())
	http.Redirect(w, r, r.URL.Path, http.StatusFound)
	return true
}
