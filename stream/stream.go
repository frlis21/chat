// Simple event broker (ab)using channels.
package stream

import "sync"

type Stream[T any] struct {
	mu    sync.Mutex
	chans map[chan T]struct{}
}

func New[T any]() *Stream[T] {
	return &Stream[T]{
		chans: make(map[chan T]struct{}),
	}
}

func (s *Stream[T]) Chan(n uint) chan T {
	s.mu.Lock()
	ch := make(chan T, n)
	s.chans[ch] = struct{}{}
	s.mu.Unlock()
	return ch
}

func (s *Stream[T]) Close(ch chan T) {
	s.mu.Lock()
	delete(s.chans, ch)
	s.mu.Unlock()
	close(ch)
}

func (s *Stream[T]) Wrap(ch chan T) {
	for v := range ch {
		s.Receive(v)
	}
}

func (s *Stream[T]) Receive(v T) {
	s.mu.Lock()
	chans := make([]chan T, 0, len(s.chans))
	for ch, _ := range s.chans {
		chans = append(chans, ch)
	}
	s.mu.Unlock()
	for _, ch := range chans {
		ch <- v
	}
}
