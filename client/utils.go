package main

import (
	"chat/client/client"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type GroupPageContent struct {
	Group    *client.Group
	Messages []*client.Message
}

type HomePageContent struct {
	User   *client.User
	Groups []*client.Group
}

type RelayPageContent struct {
	Relays []*client.Relay
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
	user, err := client.GetCurrentUser()
	if err != nil {
		return GroupPageContent{}, errors.New(client.MISSING_USER)
	}
	m := client.NewMessage(fmt.Sprintf("%v", g), g.Antecedent, content, time.Now(), user)
	err = g.SendMessage(m)
	return GroupPageContent{g, g.GetMessages()}, err
}

func userSetup(req *http.Request) error {
	username := req.FormValue("username")
	err := client.SetUser(username)
	if err != nil {
		return err
	}
	return nil
}

func addRelay(req *http.Request) error {
	address := req.FormValue("address")
	port, err := strconv.Atoi(req.FormValue("port"))
	if err != nil {
		return err
	}
	err = client.AddRelay(address, port)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	return nil
}

func searchGroups(req *http.Request) []*client.Group {
	// for _, relay := range client.GetRelays() {
	// 	client.
	// }
	return []*client.Group{}
}
