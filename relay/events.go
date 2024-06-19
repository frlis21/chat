package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strconv"
)

import (
	"chat/sse"
	"chat/stream"
)

type SubscriptionRequest struct {
	UserID string   `json:"user_id"`
	Topics []string `json:"topics"`
}

type EventServer struct {
	DB     *PostStore
	Online *Presence
	Posts  *stream.Stream[Post]
}

func (s *EventServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Parse the subscription request
	dec := json.NewDecoder(req.Body)
	var subreq SubscriptionRequest
	if err := dec.Decode(&subreq); err != nil {
		http.Error(w, "Could not read JSON body", http.StatusBadRequest)
		return
	}
	if len(subreq.Topics) == 0 {
		http.Error(w, "Must specify at least 1 topic", http.StatusBadRequest)
		return
	}

	// Register user for presence.
	// User is removed when context reports done.
	s.Online.Add(req.Context(), subreq.Topics, subreq.UserID)

	// Note: This implicitly "pauses" the stream (globally) until `ch` is read from.
	// We might want some buffering if getting replay history is slow.
	ch := s.Posts.Chan()
	ctx := req.Context()
	context.AfterFunc(ctx, func() {
		s.Posts.Close(ch)
	})

	events := stream.New[sse.Event](32)
	go sse.New(events).ServeHTTP(w, req)

	// Get posts to replay
	lastEventId, _ := strconv.ParseUint(req.Header.Get("Last-Event-ID"), 10, 64)
	var replay []Post
	for _, k := range s.DB.Keys() {
		if post, ok := s.DB.Get(k); ok && post.Sequence > lastEventId && slices.Contains(subreq.Topics, post.Topic) {
			replay = append(replay, post)
		}
	}
	slices.SortFunc(replay, func(a, b Post) int {
		return int(a.Sequence - b.Sequence)
	})

	// Replay posts.
	for _, p := range replay {
		data, err := json.Marshal(p)
		if err != nil {
			log.Println("Bad post in DB!")
			continue
		}
		events.Receive(sse.Event{
			Name: "post",
			Data: data,
			ID:   strconv.FormatUint(p.Sequence, 10),
		})
	}

	// Filter out posts we are not subscribed to
	for p := range ch {
		if slices.Contains(subreq.Topics, p.Topic) {
			data, err := json.Marshal(p)
			if err != nil {
				log.Println("Bad post in DB!")
				continue
			}
			events.Receive(sse.Event{
				Name: "post",
				Data: data,
				ID:   strconv.FormatUint(p.Sequence, 10),
			})
		}
	}
}
