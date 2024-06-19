package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"chat/client/client"
)

type GroupPageContent struct {
	Group    *client.Group
	Messages []*client.Message
}

type HomePageContent struct {
	User           *client.User
	Groups         []*client.Group
	SearchedGroups []*client.Group
}

type RelayPageContent struct {
	Relays []*client.Relay
}

var groups map[string]*client.Group = make(map[string]*client.Group)
var searchedGroups map[string]*client.Group = make(map[string]*client.Group)

func homepage(w http.ResponseWriter, req *http.Request) {
	var foundGroups []*client.Group = nil
	if req.Method == "POST" {
		foundGroups = client.SearchGroups(req)
		fmt.Printf("%v\n", foundGroups)
		for _, group := range foundGroups {
			searchedGroups[group.UUID] = group
		}
	}
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error loading page\n"))
		return
	}
	if len(groups) == 0 {
		groups = client.GetGroups()
	}
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
	pageContent := HomePageContent{user, data, foundGroups}
	tmpl.Execute(w, pageContent)
}

func viewGroup(w http.ResponseWriter, req *http.Request) {
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
		// fmt.Printf("%v\n", g)
		// messages := g.GetMessages()
		// fmt.Printf("%v\n", messages[0].Author.UUID)
		page = GroupPageContent{g, g.GetMessages()}
	} else if req.Method == "POST" {
		content := req.FormValue("message")
		// fmt.Printf("%v\n", content)
		user, err := client.GetCurrentUser()
		if err != nil {
			http.Redirect(w, req, "/setup", http.StatusSeeOther)
		}
		m := client.NewMessage(g.UUID, g.Antecedent, content, time.Now(), user)
		err = g.SendMessage(m)
		if err != nil {
			fmt.Printf("%v\n", err)
			w.Write([]byte("Error sending message\n"))
			return
		}
		page = GroupPageContent{g, g.GetMessages()}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Invalid Request Method"))
		return
	}

	tmpl.Execute(w, page)
}

func joinGroup(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("UUID")
	err := uuid.Validate(id)
	if err != nil {
		w.Write([]byte("Invalid Group ID\n"))
		return
	}
	g := searchedGroups[id]
	err = g.JoinGroup()
	if err != nil {
		w.Write([]byte("Unable to join group\n"))
	}
	groups[g.UUID] = g
	delete(searchedGroups, g.UUID)
	http.Redirect(w, req, fmt.Sprintf("/group/view/%v", id), http.StatusSeeOther)

}

func createGroup(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" || req.Method == "" {
		tmpl, err := template.ParseFiles("templates/create_group.html")
		if err != nil {
			fmt.Printf("%v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error loading page\n"))
			return
		}
		tmpl.Execute(w, nil)
		return
	} else if req.Method == "POST" {
		groupName := req.FormValue("group_name")
		g := client.CreateGroup(groupName)
		if g == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error creating group\n"))
			return
		}
		groups[g.UUID] = g
		http.Redirect(w, req, "/", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Invalid Request Method"))
		return
	}
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
		username := req.FormValue("username")
		err := client.SetUser(username)
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
		address := req.FormValue("address")
		port, err := strconv.Atoi(req.FormValue("port"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed creating relay"))
			return
		}
		err = client.AddRelay(address, port)
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
	port := ":" + os.Args[1]
	http.HandleFunc("/group/view/{UUID}", viewGroup)
	http.HandleFunc("/group/join/{UUID}", joinGroup)
	http.HandleFunc("/group/create", createGroup)
	http.HandleFunc("/setup", initialSetup)
	http.HandleFunc("/relay", setupRelay)
	http.HandleFunc("/", homepage)
	http.ListenAndServe(port, nil)
}
