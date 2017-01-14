package main

import (
	"html/template"
	"time"
	"net/http"

	"github.com/7joe7/hamrchecker/web"
	"github.com/7joe7/hamrchecker/db"
	"github.com/7joe7/hamrchecker/checker"
)

func main() {
	templates, err := template.ParseFiles("web/resources/html/hamrchecker.html")
	if err != nil {
		panic(err)
	}
	web.SetTemplates(templates)
	http.HandleFunc("/", web.Index)
	http.Handle("/web/resources/", http.StripPrefix("/web/resources/", http.FileServer(http.Dir("/usr/local/src/web/resources"))))

	searches := db.GetSearches()
	for i := 0; i < len(searches); i++ {
		if time.Now().After(*searches[i].Till) {
			db.RemoveSearch(searches[i].Id)
		} else {
			go checker.RunSearch(searches[i])
		}
	}
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
