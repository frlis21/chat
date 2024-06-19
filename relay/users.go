package main

import (
	"encoding/json"
	"net/http"
)

type UserServer struct {
	db *PostStore
	presence *Presence
}

func (s *UserServer) queryOnline(w http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	if !values.Has("topic") {
		http.Error(w, "Missing topic", http.StatusBadRequest)
		return
	}
	results := s.presence.Search(values.Get("topic"), values.Get("name"))
	enc := json.NewEncoder(w)
	if err := enc.Encode(results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//func (s *UserServer) queryAll(w http.ResponseWriter, req *http.Request) {
//	values := req.URL.Query()
//}

func (s *UserServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.queryOnline(w, req)
	//values := req.URL.Query()
	//if values.Has("online") {
	//	s.queryOnline(w, req)
	//} else
	//	s.queryAll(w, req)
	//}
}


