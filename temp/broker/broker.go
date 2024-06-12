// Simple event broker
// a.k.a. one-to-many channel, broadcast, etc.

package broker

type Sub[T any] struct {
	topic string
	sub   chan T
}

type Pub[T any] struct {
	topic string
	msg   T
}

type Broker[T any, M any] struct {
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

func (b *Broker[T, M]) Subscribe(topic T, ch chan M) {
	b.sub <- Sub[T]{topic, ch}
}

func (b *Broker[T]) Publish(topic T, msg M) {
	b.pub <- Pub[T]{topic, msg}
}

func (b *Broker[T]) Unsubscribe(topic T, ch chan M) {
	b.unsub <- Sub[T]{topic, ch}
}

