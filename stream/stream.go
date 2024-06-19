// Simple event broker (ab)using channels.
package stream

import "sync"

type Stream[T any] struct {
	buf   uint
	mu    sync.Mutex
	chans map[chan T]struct{}
}

func New[T any](n uint) *Stream[T] {
	return &Stream[T]{
		buf: n,
		chans: make(map[chan T]struct{}),
	}
}

func (s *Stream[T]) Chan() chan T {
	s.mu.Lock()
	ch := make(chan T, s.buf)
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
	var chans []chan T
	s.mu.Lock()
	for ch, _ := range s.chans {
		chans = append(chans, ch)
	}
	s.mu.Unlock()
	for _, ch := range chans {
		ch <- v
	}
}
