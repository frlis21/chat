package main

import (
	//"crypto/sha256"
	//"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"sync/atomic"
	"time"
)

import (
	"chat/stream"
)

type UIDGen struct {
	atomic.Uint64
}

func (i *UIDGen) Next() uint64 {
	return i.Add(1)
}

type PublishServer struct {
	DB     *PostStore
	Client *http.Client
	Posts  *stream.Stream[Post]
	UID    *UIDGen
}

type PostRequest struct {
	// Client-calculated ID
	ID string `json:"id"`
	// ID of parent post
	Parent string `json:"parent"`
	// ID of genesis post
	Topic string `json:"topic"`
	// User ID
	Author string `json:"author"`
	// Post content
	Content string `json:"content"`
	// Client timestamp
	CreatedAt time.Time `json:"created_at"`
	// Relays to contact in case this relay doesn't have full history
	Relays []string `json:"relays"`
	// In a real app clients would sign messages.
	// Here the authentication is nonexistent
	// and spam/impersonation would be a problem
	// in a Byzantine failure model.
}

func (s *PublishServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var postreq PostRequest
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&postreq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hash sanity check
	//ts, _ := postreq.CreatedAt.MarshalBinary()
	//h := sha256.New()
	//h.Write([]byte(postreq.Parent))
	//h.Write([]byte(postreq.Author))
	//h.Write([]byte(postreq.Content))
	//h.Write(ts)
	//id := hex.EncodeToString(h.Sum(nil))

	//if id != postreq.ID {
	//	http.Error(w, "Check your hash", http.StatusBadRequest)
	//	return
	//}

	newpost := Post{
		ID:        postreq.ID,
		Topic:     postreq.Topic,
		Parent:    postreq.Parent,
		Author:    postreq.Author,
		Content:   postreq.Content,
		CreatedAt: postreq.CreatedAt,
		Sequence:  s.UID.Next(),
	}

	// Genesis post
	if newpost.ID == newpost.Topic {
		w.WriteHeader(http.StatusCreated)
		_, changed := s.DB.GetDefault(newpost.ID, newpost)
		if !changed {
			http.Error(w, "Topic exists", http.StatusConflict)
			return
		}
		s.Posts.Receive(newpost)
		return
	}

	var relays map[string]*url.URL
	for _, addr := range postreq.Relays {
		relayURL, err := url.Parse(addr)
		if err != nil || !relayURL.IsAbs() {
			continue
		}
		relays[addr] = relayURL
	}

	chain := []Post{newpost}
	_, ok := s.DB.Get(chain[len(chain)-1].Parent)

	// TODO cap length and paginate
	for length := 1; len(relays) > 0 && !ok; length *= 2 {
		//log.Printf("parent: %s\n", chain[len(chain)-1].Parent)
		for k, relay := range relays {
			v := url.Values{}
			v.Add("length", strconv.Itoa(length))
			relay = relay.JoinPath("posts", chain[len(chain)-1].Parent)
			relay.RawQuery = v.Encode()
			resp, err := s.Client.Get(relay.String())
			if err != nil {
				// This relay ran out of history
				delete(relays, k)
				continue
			}
			dec := json.NewDecoder(resp.Body)
			var segment []Post
			if err := dec.Decode(&segment); err != nil {
				// This relay is probably compromised
				delete(relays, k)
				continue
			}
			// In a Byzantine failure model
			// we would check the history and if something doesn't match
			// we reject it and mark the relay we got it from as compromised.
			chain = append(chain, segment...)
			_, ok = s.DB.Get(chain[len(chain)-1].Parent)
			break
		}
	}

	if !ok {
		http.Error(w, "Unable to reconcile history", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)

	slices.Reverse(chain)
	for _, post := range chain {
		s.DB.Put(post.ID, post)
		s.Posts.Receive(post)
	}
}
