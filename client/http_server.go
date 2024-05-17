package main

import (
	"client/client"
	"html/template"
	"net/http"
)

var groups map[string]*client.Group = nil

func homepage(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		w.Write([]byte("Error loading page\n"))
	}
	groups = client.GetGroups()
	tmpl.Execute(w, groups)
}

func group(w *http.ResponseWriter, req *http.Request) {
}

func main() {
	http.HandleFunc("/", homepage)
	http.ListenAndServe(":8080", nil)
}
