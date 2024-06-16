package main

import (
	"client/client"
	"fmt"
	"net/http"
)

type GroupPageContent struct {
	Group    *client.Group
	Messages []*client.Message
}

func groupGetHandler(g *client.Group) GroupPageContent {
	return GroupPageContent{g, g.GetMessages()}
}

func groupPostHandler(g *client.Group, req *http.Request) (GroupPageContent, error) {
	err := req.ParseForm()
	if err != nil {
		fmt.Printf("%v\n", err)
		return GroupPageContent{}, err
	}

	content := req.FormValue("message")
	// fmt.Printf("%v\n", content)
	m := client.NewMessage(content, client.NewUser("NEW USERNAME", "127.0.0.1"))
	err = g.SendMessage(m)
	return GroupPageContent{g, g.GetMessages()}, err
}
