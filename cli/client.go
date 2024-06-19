package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
	"time"

	"chat/sse"
	"chat/store"
	"chat/stream"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Topic struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Post struct {
	ID        string    `json:"id"`
	Parent    string    `json:"parent"`
	Topic     string    `json:"topic"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type PostRequest struct {
	Post
	Relays []string `json:"relays"`
}

type Relay struct {
	//json.Marshaler
	//json.Unmarshaler
	mu sync.Mutex
	Client      *http.Client
	DB          *store.Store[string, Post]
	LastEventID *atomic.Uint64
	Retry       *atomic.Uint64
	URL         *url.URL
}

type RelayJSON struct {
	LastEventID uint64 `json:"last_event_id"`
	Retry       uint64 `json:"last_event_id"`
	URL         string `json:"url"`
}

func (r *Relay) MarshalJSON() ([]byte, error) {
	m := RelayJSON{
		LastEventID: r.LastEventID.Load(),
		Retry:       r.Retry.Load(),
		URL:         r.URL.String(),
	}
	return json.Marshal(m)
}

func (r *Relay) UnmarshalJSON(data []byte) (err error) {
	var m RelayJson
	if err = json.Unmarshal(data, &m); err != nil {
		return err
	}
	r.LastEventID.Store(m.LastEventID)
	r.Retry.Store(m.Retry)
	r.URL, err = url.Parse(m.URL)
	return err
}

func (r *Relay) Publish(req PostRequest) (err error) {
	var resp *http.Response
	endpoint := r.URL.Join("posts")
	endpoint.RawQuery = query.Encode()
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err = enc.Encode(req); err != nil {
		return err
	}
	resp, err = r.Client.Post(endpoint, "application/json", buf)
	if resp.StatusCode != http.StatusCreated {
		return results, fmt.Errorf("Search error: %s", resp.Status)
	}
	return err
}

func (r *Relay) Stream(posts *stream.Stream[Post]) (err error) {
	req, err := http.NewRequest("POST", r.URL.Join("events"), "")
	req.Header.Set("Last-Event-ID", strconv.FormatUint(r.LastEventID, 10))
	resp, err := r.Client.Do(req)
	er := sse.Reader(resp.Body)
	for e, err := er.ReadEvent(); err == nil; e, err = er.ReadEvent() {
		var post Post
		if err := json.Unmarshal(e.Data, &Post); err != nil {
			// Post should really be PostOrError or something
			// so that we can stream errors.
			continue
		}
		
		posts.Receive(post)
	}
}

func Query[T any](r *Relay, path string, query url.Values) (result T, err error) {
	var resp *http.Response
	endpoint := r.URL.Join(path)
	endpoint.RawQuery = query.Encode()
	resp, err = r.Client.Get(endpoint)
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("Search error: %s", resp.Status)
	}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&result)
	return result, nil
}

func (r *Relay) SearchTopics(query url.Values) (map[string]string, err error) {
	topics, err := Query[[]Topic](r, "topics", query)
	m := make(map[string]string, 0, len(topics))
	for _, t := range topics {
		m[t.ID] = t.Name
	}
	return m
}

func (r *Relay) SearchPosts(query url.Values) (map[string]Post, err error) {
	posts, err := Query[[]Post](r, "posts", query)
	m := make(map[string]Post, 0, len(posts))
	for _, p := range posts {
		m[p.ID] = p
	}
}

func (r *Relay) SearchUsers(query url.Values) (map[string]string, err error) {
	users, err := Query[[]User](r, "users", query)
	m := make(map[string]string, 0, len(posts))
	for _, p := range posts {
		m[p.ID] = p
	}
}

// Relay multiplexer.
// Up to library user to save posts and relays persistently.
type Client struct {
	Relays []Relay
	Posts  *store.Store[string, Post]
}

func merge[K comparable, V any](a, b map[K]V) {
	//merged := make(map[K]V, 0, len(a) + len(b))
	for k, v := range b {
		a[k] = v
	}
}

// TODO implement more sophisticated error handling

func (c *Client) SearchUsers(query url.Values) map[string]string {
	all := make(map[string]string)
	for _, r := range c.Relays {
		users, err := r.SearchUsers(query)
		if err != nil {
			continue
		}
		merge(all, users)
	}
	return all
}

func (c *Client) SearchPosts(query url.Values) map[string]Post {
	all := make(map[string]Post)
	for _, r := range c.Relays {
		posts, err := r.SearchPosts(query)
		if err != nil {
			continue
		}
		merge(all, posts)
	}
	return all
}

func (c *Client) SearchTopics(query url.Values) map[string]string {
	all := make(map[string]Post)
	for _, r := range c.Relays {
		topics, err := r.SearchTopics(query)
		if err != nil {
			continue
		}
		merge(all, topics)
	}
	return all
}

func (c *Client) Stream() (posts *stream.Stream[Post], err error) {
	for _, relay := range relays {
		req, err := http.NewRequest("POST", url.JoinPath(relay, "events"), "")
		if err != nil {
			return nil, err
		}
		resp, err := c.httpc.Do(req)
	}
}
