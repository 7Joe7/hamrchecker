package web

import (
	"fmt"
	"net/http"
	"html/template"

	"github.com/7joe7/hamrchecker/resources"
	"github.com/7joe7/hamrchecker/db"
	"github.com/7joe7/hamrchecker/checker"
)

var (
	templates *template.Template
)

func indexPost(w http.ResponseWriter, r *http.Request, ti *resources.TemplateInfo) {
	if err := r.ParseForm(); err != nil {
		ti.Flash = fmt.Sprintf("Unable to parse form. %v", err)
		ti.FlashType = "error"
		return
	}
	search := &resources.Search{}
	if err := search.Parse(
		r.PostFormValue("email"),
		r.PostFormValue("date"),
		r.PostFormValue("from"),
		r.PostFormValue("till"),
		r.PostFormValue("length"),
		r.PostFormValue("place"),
		r.PostFormValue("sport")); err != nil {
		ti.Flash = fmt.Sprintf("Invalid data. %v", err)
		ti.FlashType = "error"
		return
	}
	if int(search.Till.Sub(resources.GetTimeFromTimePointer(search.From)).Minutes()) / 30 < search.Length {
		ti.Flash = fmt.Sprintf("Invalid number of half hours. It is higher than time range.")
		ti.FlashType = "error"
	}
	db.AddSearch(search)
	go checker.RunSearch(search)
	ti.Flash = fmt.Sprintf("Search is running. You will be informed on the inserted e-mail.")
	ti.FlashType = "success"
}

func index(w http.ResponseWriter, r *http.Request) {
	ti := resources.TemplateInfo{Searches: db.GetSearches(), FlashType:"none"}
	switch r.Method {
	case "POST":
		indexPost(w, r, &ti)
	}
	templates.ExecuteTemplate(w, "hamrchecker", ti)
}

func setTemplates(t *template.Template) {
	templates = t
}
