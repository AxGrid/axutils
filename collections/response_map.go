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
	timeout int
}

func NewResponseMap[K comparable, V any]() *ResponseMapBuilder[K, V] {
	return &ResponseMapBuilder[K, V]{
		timeout: 300,
	}
}

func (b *ResponseMapBuilder[K, V]) WithTimeout(timeout int) *ResponseMapBuilder[K, V] {
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
	timeout int
	mu      sync.RWMutex
	m       map[K]*chansHolder[V]
}

type chansHolder[V any] struct {
	t     *time.Timer
	chans []chan V
}

func (r *ResponseMap[K, V]) Set(key K, value V) {
	r.mu.RLock()
	holder, ok := r.m[key]
	r.mu.RUnlock()
	if !ok {
		return
	}
	for _, ch := range holder.chans {
		ch <- value
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.m, key)
}

func (r *ResponseMap[K, V]) Wait(key K) V {
	return <-r.getChan(key)
}

func (r *ResponseMap[K, V]) getChan(key K) chan V {
	ch := make(chan V)
	r.mu.RLock()
	holder, ok := r.m[key]
	r.mu.RUnlock()
	if ok {
		r.mu.Lock()
		defer r.mu.Unlock()
		holder.chans = append(r.m[key].chans, ch)
		r.m[key] = holder
		return ch
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	holder, ok = r.m[key]
	if ok {
		holder.chans = append(r.m[key].chans, ch)
		r.m[key] = holder
		return ch
	}
	holder = &chansHolder[V]{
		t:     time.NewTimer(time.Duration(r.timeout) * time.Second),
		chans: make([]chan V, 0),
	}
	holder.chans = append(holder.chans, ch)
	r.m[key] = holder
	go func() {
		defer holder.t.Stop()
		select {
		case <-holder.t.C:
			for _, ch := range holder.chans {
				close(ch)
			}
			r.mu.Lock()
			defer r.mu.Unlock()
			delete(r.m, key)
		}
	}()
	return ch
}
