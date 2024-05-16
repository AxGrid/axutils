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

type ResponseMapBuilder[K comparable, V any] struct {
	timeout time.Duration
}

func NewResponseMap[K comparable, V any]() *ResponseMapBuilder[K, V] {
	return &ResponseMapBuilder[K, V]{
		timeout: time.Second * 100,
	}
}

func (b *ResponseMapBuilder[K, V]) WithTimeout(timeout time.Duration) *ResponseMapBuilder[K, V] {
	b.timeout = timeout
	return b
}

func (b *ResponseMapBuilder[K, V]) Build() *ResponseMap[K, V] {
	rm := &ResponseMap[K, V]{
		timeout: b.timeout,
		mu:      sync.RWMutex{},
		m:       make(map[K]*chansHolder[V]),
	}
	return rm
}

type ResponseMap[K comparable, V any] struct {
	timeout time.Duration
	mu      sync.RWMutex
	m       map[K]*chansHolder[V]
}

type chansHolder[V any] struct {
	t         *time.Timer
	mu        sync.RWMutex
	isExist   bool
	data      V
	dataCh    chan V
	listeners []chan V
}

func newChansHolder[V any](timeout time.Duration) *chansHolder[V] {
	h := &chansHolder[V]{
		t:      time.NewTimer(timeout),
		dataCh: make(chan V, 1),
	}
	go func() {
		defer h.t.Stop()
		for {
			select {
			case <-h.t.C:
				for _, ch := range h.listeners {
					close(ch)
				}
			case d := <-h.dataCh:
				h.mu.Lock()
				h.data = d
				h.isExist = true
				for _, ch := range h.listeners {
					ch <- d
				}
				h.mu.Unlock()
			}
		}
	}()
	return h
}

func (h *chansHolder[V]) set(data V) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.isExist {
		return
	}
	h.dataCh <- data
}

func (h *chansHolder[V]) wait() V {
	h.mu.RLock()
	e := h.isExist
	h.mu.RUnlock()
	if e {
		return h.data
	}
	h.mu.Lock()
	e = h.isExist
	if e {
		h.mu.Unlock()
		return h.data
	}
	ch := make(chan V, 1)
	h.listeners = append(h.listeners, ch)
	h.mu.Unlock()
	return <-ch
}

func (r *ResponseMap[K, V]) Set(key K, value V) {
	holder := r.getHolder(key)
	holder.set(value)
}

func (r *ResponseMap[K, V]) Wait(key K) V {
	holder := r.getHolder(key)
	return holder.wait()
}

func (r *ResponseMap[K, V]) getHolder(key K) *chansHolder[V] {
	r.mu.RLock()
	holder, ok := r.m[key]
	r.mu.RUnlock()
	if ok {
		return holder
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	holder, ok = r.m[key]
	if ok {
		return holder
	}
	holder = newChansHolder[V](r.timeout)
	r.m[key] = holder
	return holder
}
