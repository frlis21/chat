package main

import (
	"context"
	"strings"
	"sync"
)

// Presence service
type Presence struct {
	mu     sync.Mutex
	online map[string]map[string]int
}

func NewPresence() *Presence {
	return &Presence{
		online: make(map[string]map[string]int),
	}
}

// We need to count "presences" for concurrency reasons.
func (p *Presence) Add(ctx context.Context, topics []string, user string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// This must be done before registering AfterFunc
	for _, topic := range topics {
		users, ok := p.online[topic]
		if !ok {
			users = make(map[string]int)
			p.online[topic] = users
		}
		users[user] = users[user] + 1
	}
	// Clean up
	context.AfterFunc(ctx, func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		for _, topic := range topics {
			p.online[topic][user] = p.online[topic][user] - 1
			if p.online[topic][user] <= 0 {
				delete(p.online[topic], user)
			}
		}
	})
}

// Call this a microservice ;)
func (p *Presence) Search(topic string, prefix string) []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	results := []string{}
	for name := range p.online[topic] {
		if strings.HasPrefix(name, prefix) {
			results = append(results, name)
		}
	}
	return results
}
