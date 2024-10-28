package collections

import (
	"sync"
	"time"
)

type RequestMap[K comparable, V any] struct {
	data        map[K][]chan V
	response    map[K]V
	mu          sync.RWMutex
	deleteAfter time.Duration
}

func NewRequestMap[K comparable, V any](ttl time.Duration) *RequestMap[K, V] {
	return &RequestMap[K, V]{
		data:        make(map[K][]chan V),
		response:    make(map[K]V),
		mu:          sync.RWMutex{},
		deleteAfter: ttl,
	}
}

func (rm *RequestMap[K, V]) GetOrCreate(key K, f func(k K) V) V {
	rm.mu.RLock()
	v, ok := rm.response[key]
	rm.mu.RUnlock()
	if ok {
		return v
	}
	rm.mu.Lock()
	v, ok = rm.response[key]
	if ok {
		rm.mu.Unlock()
		return v
	}
	ch := make(chan V, 1)
	_, ok = rm.data[key]
	rm.data[key] = append(rm.data[key], ch)
	rm.mu.Unlock()
	if !ok {
		go func() {
			vx := f(key)
			rm.mu.Lock()
			rm.response[key] = vx
			rm.mu.Unlock()
			rm.mu.RLock()
			for _, c := range rm.data[key] {
				c <- vx
			}
			rm.mu.RUnlock()
			rm.mu.Lock()
			delete(rm.data, key)
			rm.mu.Unlock()
			// Start remove goroutine
			go func() {
				time.Sleep(rm.deleteAfter)
				rm.mu.Lock()
				defer rm.mu.Unlock()
				delete(rm.data, key)
				delete(rm.response, key)
			}()
		}()
	}
	return <-ch
}
