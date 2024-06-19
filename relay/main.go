package main

import (
	"log"
	"net"
	"net/http"
	"time"
)

import (
	"chat/store"
	"chat/stream"
)

type Post struct {
	ID        string    `json:"id"`
	Parent    string    `json:"parent"`
	Topic     string    `json:"topic"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Sequence  uint64    `json:"-"`
}

type PostStore = store.Store[string, Post]

//func NewPostStore() *PostStore {
//	return PostStore(store.New[string, Post]())
//}

func main() {
	// Need to start our own listener for dynamic port
	// (as opposed to using ListenAndServe)
	listener, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on", "http://"+listener.Addr().String())

	// TODO read configuration or flags for propagation relays
	//relays := []string{}

	// Our "database"
	db := store.New[string, Post]()
	// Presence "microservice"
	presence := NewPresence()
	// Sequence number generation for SSE event IDs
	var uidGen UIDGen
	// One-many Post broadcast
	posts := stream.New[Post](32)

	// HTTP client required by PublishServer to reconcile history
	client := http.Client{
		// Set a relatively short timeout so we fail fast
		Timeout: time.Second * 1,
	}

	mux := http.NewServeMux()

	// Subscription request
	mux.Handle("POST /events", &EventServer{db, presence, posts})
	// Create a new (possibly initial) post
	mux.Handle("POST /posts", &PublishServer{db, &client, posts, &uidGen})

	// Query post history
	mux.Handle("GET /posts", &PostServer{db})
	// Get post chain starting from `id`
	mux.Handle("GET /posts/{id}", &ChainServer{db})

	// Query users
	mux.Handle("GET /users", &UserServer{db, presence})
	// Query topics
	mux.Handle("GET /topics", &TopicServer{db})

	// XXX: Does not support HTTP/2 :T
	// Dunno if this is a problem...
	log.Fatal(http.Serve(listener, mux))
}
