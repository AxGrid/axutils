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
	t         *time.Timer
	mu        sync.RWMutex
	writer    chan V
	listeners []chan V
}

func (h *chansHolder[V]) addListener(listener chan V) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners = append(h.listeners, listener)
}

func (r *ResponseMap[K, V]) Set(key K, value V) {
	r.mu.RLock()
	holder, ok := r.m[key]
	r.mu.RUnlock()
	if ok {
		for _, ch := range holder.listeners {
			ch <- value
		}
		r.mu.Lock()
		delete(r.m, key)
		r.mu.Unlock()
		return
	}
	holder = &chansHolder[V]{
		t:      time.NewTimer(time.Duration(r.timeout) * time.Second),
		mu:     sync.RWMutex{},
		writer: make(chan V, 1),
	}
	go func() {
		defer holder.t.Stop()
		for {
			select {
			case <-holder.t.C:
				holder.mu.Lock()
				for _, ch := range holder.listeners {
					close(ch)
				}
				r.mu.Lock()
				delete(r.m, key)
				r.mu.Unlock()
				holder.mu.Unlock()
			}
		}
	}()
	holder.writer <- value
	holder.mu.Lock()
	r.m[key] = holder
	holder.mu.Unlock()
}

func (r *ResponseMap[K, V]) Wait(key K) V {
	return <-r.getChan(key)
}

func (r *ResponseMap[K, V]) getChan(key K) chan V {
	r.mu.RLock()
	holder, ok := r.m[key]
	r.mu.RUnlock()
	if ok {
		ch := make(chan V, 1)
		if holder.listeners == nil {
			msg := <-holder.writer
			ch <- msg
			return ch
		}
		holder.addListener(ch)
		return ch
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	holder, ok = r.m[key]
	if ok {
		ch := make(chan V, 1)
		if holder.listeners == nil {
			msg := <-holder.writer
			ch <- msg
			return ch
		}
		holder.addListener(ch)
		return ch
	}
	holder = &chansHolder[V]{
		t:  time.NewTimer(time.Duration(r.timeout) * time.Second),
		mu: sync.RWMutex{},
	}
	go func() {
		defer holder.t.Stop()
		for {
			select {
			case <-holder.t.C:
				holder.mu.Lock()
				for _, ch := range holder.listeners {
					close(ch)
				}
				r.mu.Lock()
				delete(r.m, key)
				r.mu.Unlock()
				holder.mu.Unlock()
			}
		}
	}()
	ch := make(chan V, 1)
	holder.addListener(ch)
	r.m[key] = holder
	return ch
}

//func (r *ResponseMap[K, V]) Set(key K, value V) {
//	r.mu.RLock()
//	holder, ok := r.m[key]
//	r.mu.RUnlock()
//	if !ok {
//		return
//	}
//	for _, ch := range holder.listeners {
//		ch <- value
//	}
//	r.mu.Lock()
//	defer r.mu.Unlock()
//	delete(r.m, key)
//}
//
//func (r *ResponseMap[K, V]) Wait(key K) V {
//	return <-r.getChan(key)
//}
//
//func (r *ResponseMap[K, V]) getChan(key K) chan V {
//	ch := make(chan V)
//	r.mu.RLock()
//	holder, ok := r.m[key]
//	r.mu.RUnlock()
//	if ok {
//		r.mu.Lock()
//		defer r.mu.Unlock()
//		holder.listeners = append(r.m[key].listeners, ch)
//		r.m[key] = holder
//		return ch
//	}
//	r.mu.Lock()
//	defer r.mu.Unlock()
//	holder, ok = r.m[key]
//	if ok {
//		holder.listeners = append(r.m[key].listeners, ch)
//		r.m[key] = holder
//		return ch
//	}
//	holder = &chansHolder[V]{
//		t:         time.NewTimer(time.Duration(r.timeout) * time.Second),
//		listeners: make([]chan V, 0),
//	}
//	holder.listeners = append(holder.listeners, ch)
//	r.m[key] = holder
//	go func() {
//		defer holder.t.Stop()
//		select {
//		case <-holder.t.C:
//			for _, ch := range holder.listeners {
//				close(ch)
//			}
//			r.mu.Lock()
//			defer r.mu.Unlock()
//			delete(r.m, key)
//		}
//	}()
//	return ch
//}
