package collections

import (
	"github.com/go-errors/errors"
	"sync"
	"time"
)

var ErrTimeout = errors.New("timeout")

type resultHolder[V any] struct {
	result V
	err    error
}
type RequestMap[K comparable, V any] struct {
	waiters     map[K][]chan resultHolder[V]
	response    map[K]resultHolder[V]
	mu          sync.RWMutex
	deleteAfter time.Duration
}

func NewRequestMap[K comparable, V any](ttl time.Duration) *RequestMap[K, V] {
	return &RequestMap[K, V]{
		waiters:     make(map[K][]chan resultHolder[V]),
		response:    make(map[K]resultHolder[V]),
		mu:          sync.RWMutex{},
		deleteAfter: ttl,
	}
}

// GetOrCreate returns the value for the key if it exists, otherwise it calls the function f and returns the result
func (rm *RequestMap[K, V]) GetOrCreate(key K, f func(k K) V) V {
	rm.mu.RLock()
	v, ok := rm.response[key]
	rm.mu.RUnlock()
	if ok {
		return v.result
	}
	rm.mu.Lock()
	v, ok = rm.response[key]
	if ok {
		rm.mu.Unlock()
		return v.result
	}
	ch := make(chan resultHolder[V], 1)
	_, ok = rm.waiters[key]
	rm.waiters[key] = append(rm.waiters[key], ch)
	rm.mu.Unlock()
	if !ok { // Start goroutine
		go func() {
			vx := f(key)
			rm.mu.Lock()
			r := resultHolder[V]{result: vx}
			rm.response[key] = r
			rm.mu.Unlock()
			rm.mu.RLock()
			for _, c := range rm.waiters[key] {
				c <- r
			}
			rm.mu.RUnlock()
			rm.mu.Lock()
			delete(rm.waiters, key)
			rm.mu.Unlock()
			// Start remove goroutine
			go func() {
				time.Sleep(rm.deleteAfter)
				rm.mu.Lock()
				defer rm.mu.Unlock()
				delete(rm.waiters, key)
				delete(rm.response, key)
			}()
		}()
	}
	res := <-ch
	return res.result
}

// GetOrCreateWithErr returns the value for the key if it exists, otherwise it calls the function f and returns the result
func (rm *RequestMap[K, V]) GetOrCreateWithErr(key K, f func(k K) (V, error)) (V, error) {
	rm.mu.RLock()
	v, ok := rm.response[key]
	rm.mu.RUnlock()
	if ok {
		return v.result, v.err
	}
	rm.mu.Lock()
	v, ok = rm.response[key]
	if ok {
		rm.mu.Unlock()
		return v.result, v.err
	}
	ch := make(chan resultHolder[V], 1)
	_, ok = rm.waiters[key]
	rm.waiters[key] = append(rm.waiters[key], ch)
	rm.mu.Unlock()
	if !ok { // Start goroutine
		go func() {
			vx, err := f(key)
			rm.mu.Lock()
			r := resultHolder[V]{result: vx, err: err}
			rm.response[key] = r
			rm.mu.Unlock()
			rm.mu.RLock()
			for _, c := range rm.waiters[key] {
				c <- r
			}
			rm.mu.RUnlock()
			rm.mu.Lock()
			delete(rm.waiters, key)
			rm.mu.Unlock()
			// Start remove goroutine
			go func() {
				time.Sleep(rm.deleteAfter)
				rm.mu.Lock()
				defer rm.mu.Unlock()
				delete(rm.waiters, key)
				delete(rm.response, key)
			}()
		}()
	}
	res := <-ch
	return res.result, res.err
}

// Timeout returns a function that will call f with k and return the result or ErrTimeout if the duration is exceeded
func (rm *RequestMap[K, V]) Timeout(duration time.Duration, f func(k K) (V, error)) func(k K) (V, error) {
	return func(k K) (V, error) {
		var result V
		res := make(chan resultHolder[V], 1)
		defer close(res)
		timeOut := false
		go func() {
			v, err := f(k)
			if !timeOut {
				res <- resultHolder[V]{result: v, err: err}
			}
		}()
		select {
		case r := <-res:
			return r.result, r.err
		case <-time.After(duration):
			timeOut = true
			return result, ErrTimeout
		}
	}
}
