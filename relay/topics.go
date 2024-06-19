package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type TopicResponse struct {
	ID string `json:"id"`
	Name  string `json:"name"`
}

type TopicServer struct {
	db *PostStore
}

// Search for topics
func (s *TopicServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	name := values.Get("name")
	seen := make(map[string]struct{})
	topics := make([]TopicResponse, 0, 32)

	// This seems super slow, would like to profile.
	for _, k := range s.db.Keys() {
		post, ok := s.db.Get(k)
		if !ok {
			continue
		} else if _, ok := seen[post.ID]; ok {
			continue
		}
		init, ok := s.db.Get(post.Topic)
		if ok && strings.HasPrefix(init.Content, name) {
			seen[post.ID] = struct{}{}
			// Remember: topics are "genesis blocks"
			// and topic names are that block's content.
			topics = append(topics, TopicResponse{
				ID: post.ID,
				Name:  init.Content,
			})
		}
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(topics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
