package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type ChainServer struct {
	DB *PostStore
}

func (s *ChainServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	query := req.URL.Query()
	length := 4096
	if query.Has("length") {
		var err error
		// Should probably limit but big number go brrrr fun
		if length, err = strconv.Atoi(query.Get("length")); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	chain := make([]Post, 0, length)

	for i := 0; i < length; i++ {
		p, ok := s.DB.Get(id)
		if !ok || id == p.Parent {
			break
		}
		chain = append(chain, p)
		id = p.Parent
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(chain); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
