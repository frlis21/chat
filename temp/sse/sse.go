package sse

import "net/http"

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

type Sink struct {
	unsub chan Sub[T]
	sub   chan Sub[T]
	pub   chan Pub[T]
}

type Source struct {
	unsub chan Sub[T]
	sub   chan Sub[T]
	pub   chan Pub[T]
}



// Post events
func (h *Sink) ServeHTTP(w ResponseWriter, r *Request) {
	
}

// Stream events
func (h *Source) ServeHTTP(w ResponseWriter, r *Request) {
	
}

func New() (Sink, Source) {

}
