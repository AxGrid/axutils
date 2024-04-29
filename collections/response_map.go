package collections

import "sync"

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (11.04.2024)
*/

type ResponseMap[K comparable, V any] struct {
	m  map[K]chan V
	mu sync.RWMutex
}

func (r *ResponseMap[K, V]) getChan(key K) chan V {
	r.mu.RLock()
	ch, ok := r.m[key]
	r.mu.RUnlock()
	if ok {
		return ch
	}
	r.mu.Lock()
	ch, ok = r.m[key]
	if ok {
		return ch
	}
	ch = make(chan V, 1)
	r.m[key] = ch
	return ch
}

func (r *ResponseMap[K, V]) Set(key K, value V) {
	r.getChan(key) <- value
}

func (r *ResponseMap[K, V]) Wait(key K) V {
	res := <-r.getChan(key)
	r.mu.Lock()
	delete(r.m, key)
	r.mu.Unlock()
	return res
}
