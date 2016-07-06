package main

import (
	"fmt"
	"net/http"
	"html/template"
)

func indexPost(w http.ResponseWriter, r *http.Request, ti *templateInfo) {
	if err := r.ParseForm(); err != nil {
		ti.Flash = fmt.Sprintf("Unable to parse form. %v", err)
		ti.FlashType = "error"
		return
	}
	search := &search{}
	if err := search.parse(
		r.PostFormValue("email"),
		r.PostFormValue("date"),
		r.PostFormValue("from"),
		r.PostFormValue("till"),
		r.PostFormValue("length")); err != nil {
		ti.Flash = fmt.Sprintf("Invalid data. %v", err)
		ti.FlashType = "error"
		return
	}
	addSearch(search)
	go runSearch(search)
	ti.Flash = fmt.Sprintf("Search is running. You will be informed on the inserted e-mail.")
	ti.FlashType = "success"
}

func index(w http.ResponseWriter, r *http.Request) {
	ti := templateInfo{Searches:searches, FlashType:"none"}
	switch r.Method {
	case "POST":
		indexPost(w, r, &ti)
	}
	template := template.Must(template.ParseFiles("resources/html/hamrchecker.html"))
	template.ExecuteTemplate(w, "hamrchecker", ti)
}
