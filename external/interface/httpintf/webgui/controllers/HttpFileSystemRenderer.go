package controllers

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

func NewHttpFileSystemRenderer(fs http.FileSystem) (*HttpFileSystemRenderer, error) {
	renderer := &HttpFileSystemRenderer{FileSystem: fs}
	var err error
	renderer.Layout, err = renderer.fsString(`/layout.html`)
	return renderer, err
}

type HttpFileSystemRenderer struct {
	http.FileSystem
	Layout string
}

func (r *HttpFileSystemRenderer) fsString(name string) (string, error) {
	f, err := r.FileSystem.Open(name)
	if err != nil {
		return ``, err
	}
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return ``, err
	}
	return string(bs), nil
}

//TODO: cache templates if needed
func (r *HttpFileSystemRenderer) Render(w http.ResponseWriter, tempName string, data interface{}) {
	pageRawStr, err := r.fsString(tempName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.New(``)

	if tmpl, err = tmpl.New(`page`).Parse(r.Layout); err != nil {
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
