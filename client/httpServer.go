package main

import (
	"client/client"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/google/uuid"
)

var groups map[string]*client.Group = nil

func homepage(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error loading page\n"))
		return
	}
	groups = client.GetGroups()
	data := []*client.Group{}
	for _, ptr := range groups {
		data = append(data, ptr)
	}
	// fmt.Printf("%v\n", data)
	user, err := client.GetCurrentUser()
	if err != nil {
		http.Redirect(w, req, "/setup", http.StatusSeeOther)
		return
	}
	pageContent := HomePageContent{user, data}
	tmpl.Execute(w, pageContent)
}

func group(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("templates/group.html")
	if err != nil {
		fmt.Printf("%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
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
	if req.Method == "GET" || req.Method == "" {
		page = groupGetHandler(g)
	} else if req.Method == "POST" {
		page, err = groupPostHandler(g, req)
		if err != nil {
			if err.Error() == client.MISSING_USER {
				http.Redirect(w, req, "/setup", http.StatusSeeOther)
			} else {
				fmt.Printf("%v\n", err)
				w.Write([]byte("Error sending message\n"))
			}
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Invalid Request Method"))
		return
	}

	tmpl.Execute(w, page)
}

func initialSetup(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" || req.Method == "" {
		tmpl, err := template.ParseFiles("templates/initial_setup.html")
		if err != nil {
			fmt.Printf("%v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error loading page\n"))
			return
		}
		tmpl.Execute(w, nil)
	} else if req.Method == "POST" {
		err := userSetup(req)
		if err != nil {
			fmt.Printf("%v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed user creation"))
			return
		}
		http.Redirect(w, req, "/", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Invalid Request Method"))
		return
	}
}

func setupRelay(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFiles("templates/relay_setup.html")
	if err != nil {
		fmt.Printf("%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error loading page\n"))
		return
	}
	page := RelayPageContent{}
	if req.Method == "GET" || req.Method == "" {
		page.Relays = client.GetRelays()
	} else if req.Method == "POST" {
		err := addRelay(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed creating relay"))
			return
		}
		page.Relays = client.GetRelays()
		http.Redirect(w, req, "/relay", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Invalid Request Method"))
		return
	}
	tmpl.Execute(w, page)
}

func main() {
	err := os.MkdirAll(client.BASE_DATA_PATH, 0777)
	if err != nil {
		panic(fmt.Sprintf("Unable to create directory for data storage: %v", err))
	}
	http.HandleFunc("/group/{UUID}", group)
	http.HandleFunc("/setup", initialSetup)
	http.HandleFunc("/relay", setupRelay)
	http.HandleFunc("/", homepage)
	http.ListenAndServe(":8080", nil)
}
