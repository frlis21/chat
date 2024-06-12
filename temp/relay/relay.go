package main

import "log"
import "net"
import "net/http"
import "encoding/json"

type Post struct {
	Author, Body string
}

type Sub[T any] struct {
	topic string
	sub   chan T
}

type Pub[T any] struct {
	topic string
	msg   T
}

type Broker[T any] struct {
	unsub chan Sub[T]
	sub   chan Sub[T]
	pub   chan Pub[T]
}

func NewBroker[T any]() *Broker[T] {
	return new(Broker[T])
}

func (b *Broker[T]) Start() {
	subs := map[string]map[chan T]struct{}{}
	for {
		select {
		case sub := <-b.sub:
			subs[sub.topic][sub.sub] = struct{}{}
		case sub := <-b.unsub:
			delete(subs[sub.topic], sub.sub)
		case pub := <-b.pub:
			for sub := range subs[pub.topic] {
				sub <- pub.msg
			}
		}
	}
}

func (b *Broker[T]) Subscribe(topic string, ch chan T) {
	b.sub <- Sub[T]{topic, ch}
}

func (b *Broker[T]) Publish(topic string, msg T) {
	b.pub <- Pub[T]{topic, msg}
}

func (b *Broker[T]) Unsubscribe(topic string, ch chan T) {
	b.unsub <- Sub[T]{topic, ch}
}

func main() {
	// Need to start our own listener for dynamic port
	// (as opposed to using ListenAndServe)
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	log.Println("Listening on", "http://"+listener.Addr().String())

	mux := http.NewServeMux()

	//b := NewBroker[Post]()
	//go b.Start()

	mux.HandleFunc("GET /posts/{topic}", func(w http.ResponseWriter, req *http.Request) {
		// Create a new topic
		topic := req.PathValue("topic")

		// Parse URL query
		if from := req.FormValue("from"), from != "" {
			to := req.FormValue("to")
			// TODO respond with history
			return
		}

		// Event stream headers
		//w.Header().Set("Content-Type", "text/event-stream")
		//w.Header().Set("Cache-Control", "no-cache")
		//w.Header().Set("Connection", "keep-alive")

		w := sse.New(w)
		w.Message("hi")
		w.Publish("topic", "message")
	})

	mux.HandleFunc("POST /posts/{topic}", func(w http.ResponseWriter, req *http.Request) {
		topic := req.PathValue("topic")

		log.Print("New post on", topic)
		var msg Post
		dec := json.NewDecoder(req.Body)
		err = dec.Decode(&msg)
		if err != nil {
			log.Print(err)
			return
		}
		log.Printf("%s: %s", msg.Author, msg.Body)

		b.Publish(topic, msg)
		w.WriteHeader(http.StatusCreated)
		w.Close()
	})

	log.Fatal(http.Serve(listener, mux))
}
