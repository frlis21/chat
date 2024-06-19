// Interactive client CLI
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/term"

	"client"
)

type Data struct {
	Relays []Relay `json:"relays"`
	Topics map[string]string `json:"topics"`
}

func readData() (data Data) {
	f, err := os.OpenFile("data.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	var data Data
	dec := json.Decoder(f)
	if err := dec.Decode(&data); err != nil {
		log.Printf("Error reading data: %s\n", err.Error())
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	return data
}

func readHistory() *store.Store[string, Post] {
	f, err := os.OpenFile("history.json", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	var history []Post
	dec := json.Decoder(f)
	if err := dec.Decode(&config); err != nil {
		log.Printf("Error reading history: %s\n", err.Error())
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	m := make(map[string]Post, 0, len(history))
	for _, p := range history {
		m[p.ID] = p
	}
	return &store.FromMap(m)
}

// TODO write config and history

type App struct {
	t *term.Terminal
	chat *Client
	topic string
	ch chan []byte
}

// Possibly need to wrap chan
func (app *App) Write(p []byte) (n int, err error) {
	app.ch <- p
	return len(p), nil
}

func (app *App) connect(addr string) {
	r, err := url.Parse(addr)
	if err != nil || !r.IsAbs() {
		fmt.Fprintln(app, "Bad address")
		return
	}

}

func (app *App) joinTopic(name string) {
	app.topic = name
	app.t.SetPrompt(name + "> ")
}

func (app *App) createTopic(name string) {
	topics[args] = uuid.New().String()
	fmt.Fprintf(app, "Created topic %s -- %s\n", args, topics[args])
}

func (app *App) poof() {
}

func (app *App) searchGroups(name string) {
	query := url.Values{}
	query.Add("name", name)
	topics := app.chat.searchTopics(query)
	var b strings.Builder
	fmt.Fprintln(b, "Found topics:")
	for id, name := range topics {
		fmt.Fprintf(b, "%s: %s\n", id, name)
	}
	fmt.Fprint(app, b.String())
}

func (app *App) searchUsers(name string) {
	query := url.Values{}
	query.Add("name", name)
	topics := app.chat.searchUsers(query)
	var b strings.Builder
	fmt.Fprintln(b, "Found users:")
	for id, name := range topics {
		fmt.Fprintf(b, "%s: %s\n", id, name)
	}
	fmt.Fprint(app, b.String())
}

func (app *App) exec(cmd string, args string) {
	switch cmd {
	case "connect": // <address>
		app.connect(args)
	case "info":
		app.printInfo()
	case "join": // <UUID> <short name>
		id, name, _ := strings.Cut(args, " ")
		app.joinGroup(id, strings.TrimSpace(name))
	case "leave": // <short name>
		app.removeGroup(args)
	case "enter": // <short name>
		app.enterGroup(args)
	case "create": // <short name>
		app.createGroup(args)
	case "poof":
		app.poof()
	case "search":
		subcmd, name, _ := strings.Cut(args, " ")
		switch strings.TrimSpace(subcmd) {
		case "groups":
			app.searchGroups(strings.TrimSpace(name))
		case "users":
			app.searchUsers(strings.TrimSpace(name))
		default:
			fmt.Fprintf(app, "Unknown: %s\n", subcmd)
		}
	default:
		fmt.Fprintf(app, "Unknown command: %s\n", cmd)
	}
}

func (app *App) publish(msg string) {
	app.chat.Publish(app.topic, msg)
}

// TODO take context
func (app *App) servePosts() {
	posts := app.posts.Chan(0)
	for post := range posts {
		fmt.Fprintf(app, "%s: %s\n", post.Author, post.Content)
	}
}

func (app *App) showMessages() {
	for p := range app.ch {
		app.t.Write(p)
	}
}

func (app *App) start() {
	posts := stream[Post].New()
	chat := Client{client, ...}

	go app.servePosts()
	go app.showMessages()
	
	for {
		input, err := app.t.ReadLine()
		if err == io.EOF {
			break // ^C or ^D
		} else if err != nil {
			panic(err)
		}
		trimmed := strings.TrimSpace(input)
		after, isCmd := strings.CutPrefix(trimmed, "/")
		if isCmd {
			cmd, args, _ := strings.Cut(after, " ")
			app.exec(cmd, strings.TrimSpace(args))
			continue
		}
		app.publish(trimmed)
	}
}

func main() {
	// Set up console
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	t := term.NewTerminal(os.Stdin, "> ")
	log.SetOutput(t)

	client := http.Client{Timeout: 5 * time.Second}

	app := NewApp()

	app.Start()

}
