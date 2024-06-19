package main

import (
	"net/http"
	"encoding/json"
	"time"
)

type PostServer struct {
	DB *PostStore
}

func (s *PostServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	if !values.Has("topic") {
		http.Error(w, "Missing topic", http.StatusBadRequest)
		return
	}
	topic := values.Get("topic")
	// Go MarshalJSON uses RFC3339
	since, err := time.Parse(time.RFC3339, values.Get("since"))
	until, err := time.Parse(time.RFC3339, values.Get("until"))
	if err != nil {
		until = time.Now()
	}

	// This could definitely be more efficient,
	// but we need a better storage solution.
	posts := make([]Post, 0, 32)
	for _, k := range s.DB.Keys() {
		post, ok := s.DB.Get(k)
		if !ok {
			continue
		}
		if post.Topic == topic && since.Compare(post.CreatedAt) == -1 && until.Compare(post.CreatedAt) == 1 {
			posts = append(posts, post)
		}
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
