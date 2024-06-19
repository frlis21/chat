package store

// Simple generic and concurrent key-value store.
import "sync"

type Store[K comparable, V any] struct {
	mu sync.Mutex
	m  map[K]V
}

func New[K comparable, V any]() *Store[K, V] {
	return &Store[K, V]{m: make(map[K]V)}
}

func (db *Store[K, V]) Put(k K, v V) (old V, ok bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	old, ok = db.m[k]
	db.m[k] = v
	return old, ok
}

func (db *Store[K, V]) Get(k K) (val V, ok bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	val, ok = db.m[k]
	return val, ok
}

func (db *Store[K, V]) Keys() (keys []K) {
	db.mu.Lock()
	defer db.mu.Unlock()
	for k := range db.m {
		keys = append(keys, k)
	}
	return keys
}
