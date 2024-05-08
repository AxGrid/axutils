package collections

import (
	"sync"
	"time"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (11.04.2024)
*/

type ResponseMap[K comparable, V any] struct {
	timeout int
	mu      sync.RWMutex
	m       map[K]chan V
}

func NewResponseMap[K comparable, V any](timeout int) *ResponseMap[K, V] {
	return &ResponseMap[K, V]{
		timeout: timeout,
		mu:      sync.RWMutex{},
		m:       make(map[K]chan V),
	}
}

func (r *ResponseMap[K, V]) getChan(key K) chan V {
	r.mu.RLock()
	ch, ok := r.m[key]
	r.mu.RUnlock()
	if ok {
		return ch
	}
	r.mu.Lock()
	defer r.mu.Unlock()
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
	t := time.NewTimer(time.Duration(r.timeout) * time.Second)
	go func() {
		for {
			select {
			case <-t.C:
				r.mu.RLock()
				close(r.m[key])
				r.mu.RUnlock()
			}
		}
	}()
	res := <-r.getChan(key)
	r.mu.Lock()
	delete(r.m, key)
	r.mu.Unlock()
	return res
}
