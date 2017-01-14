package web

import (
	"net/http"
	"html/template"
)

func Index(w http.ResponseWriter, r *http.Request) {
	index(w, r)
}

func SetTemplates(t *template.Template) {
	setTemplates(t)
}
