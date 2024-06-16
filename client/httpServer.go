package main

import (
	"client/client"
	"fmt"
	"html/template"
	"net/http"

	"github.com/google/uuid"
)

var groups map[string]*client.Group = nil

func homepage(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		w.Write([]byte("Error loading page\n"))
	}
	groups = client.GetGroups()
	data := []*client.Group{}
	for _, ptr := range groups {
		data = append(data, ptr)
	}
	// fmt.Printf("%v\n", data)
	tmpl.Execute(w, data)
}

func group(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("templates/group.html")
	if err != nil {
		fmt.Printf("%v\n", err)
		w.Write([]byte("Error loading page\n"))
		return
	}

	id := req.PathValue("UUID")
	err = uuid.Validate(id)
	if err != nil {
		w.Write([]byte("Invalid Group ID\n"))
		return
	}

	g, ok := groups[id]
	if !ok {
		w.Write([]byte("Group not found\n"))
		return
	}

	var page GroupPageContent
	if req.Method == "" || req.Method == "GET" {
		page = groupGetHandler(g)
	} else if req.Method == "POST" {
		page, err = groupPostHandler(g, req)
		if err != nil {
			fmt.Printf("%v\n", err)
			w.Write([]byte("Error sending message\n"))
			return
		}
	} else {
		w.Write([]byte("Invalid Request Method"))
		return
	}

	tmpl.Execute(w, page)
}

func main() {
	http.HandleFunc("/group/{UUID}", group)
	http.HandleFunc("/", homepage)
	http.ListenAndServe(":8080", nil)
}
