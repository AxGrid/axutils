package collections

import (
	"context"
	"github.com/go-errors/errors"
	"sync"
	"time"
)

var ErrTimeout = errors.New("timeout")

type RequestMapInitializer[K comparable, V any] struct {
	key    K
	result V
	err    error
}

type resultHolder[K comparable, V any] struct {
	key        K
	result     V
	err        error
	resultTime time.Time
}

type RequestMap[K comparable, V any] struct {
	waiters       map[K][]chan resultHolder[K, V]
	response      map[K]resultHolder[K, V]
	mu            sync.RWMutex
	deleteAfter   time.Duration
	resultSlice   []resultHolder[K, V]
	resultSliceMu sync.RWMutex
	ctx           context.Context
}

func NewRequestMap[K comparable, V any](ctx context.Context, ttl time.Duration, init ...*RequestMapInitializer[K, V]) *RequestMap[K, V] {
	res := &RequestMap[K, V]{
		waiters:     make(map[K][]chan resultHolder[K, V]),
		response:    make(map[K]resultHolder[K, V]),
		mu:          sync.RWMutex{},
		deleteAfter: ttl,
		ctx:         ctx,
	}
	if len(init) > 0 {
		for _, i := range init {
			res.response[i.key] = resultHolder[K, V]{key: i.key, result: i.result, err: i.err, resultTime: time.Now()}
		}
	}
	go res.rmWorker()
	return res
}

func (rm *RequestMap[K, V]) rmWorker() {

	do := func() {
		rm.resultSliceMu.RLock()
		if len(rm.resultSlice) == 0 {
			rm.resultSliceMu.RUnlock()
			return
		}
		first := rm.resultSlice[0]
		rm.resultSliceMu.RUnlock()
		if time.Since(first.resultTime) < rm.deleteAfter {
			return
		}

		rm.resultSliceMu.Lock()
		defer rm.resultSliceMu.Unlock()
		rm.mu.Lock()
		defer rm.mu.Unlock()
		for i, r := range rm.resultSlice {
			if time.Since(r.resultTime) < rm.deleteAfter {
				rm.resultSlice = rm.resultSlice[i:]
				return
			}
			delete(rm.response, r.key)
			delete(rm.waiters, r.key)
		}
	}

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-time.After(time.Millisecond * 100):
			do()
		}
	}
}

func (rm *RequestMap[K, V]) Count() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	a := len(rm.waiters)
	b := len(rm.response)
	return a + b
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
	ch := make(chan resultHolder[K, V], 1)
	_, ok = rm.waiters[key]
	rm.waiters[key] = append(rm.waiters[key], ch)
	rm.mu.Unlock()
	if !ok { // Start goroutine
		go func() {
			vx := f(key)
			rm.mu.Lock()
			r := resultHolder[K, V]{key: key, result: vx, resultTime: time.Now()}
			rm.response[key] = r
			for _, c := range rm.waiters[key] {
				c <- r
			}
			delete(rm.waiters, key)
			rm.mu.Unlock()
			// Start remove goroutine
			go func() {
				//time.Sleep(rm.deleteAfter)
				//rm.mu.Lock()
				//defer rm.mu.Unlock()
				//delete(rm.waiters, key)
				//delete(rm.response, key)
				rm.resultSliceMu.Lock()
				rm.resultSlice = append(rm.resultSlice, r)
				rm.resultSliceMu.Unlock()
			}()
		}()
	}
	println("wait from chan", rm.waiters)
	res := <-ch
	println("res to chan", rm.waiters)

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
	ch := make(chan resultHolder[K, V], 1)
	_, ok = rm.waiters[key]
	rm.waiters[key] = append(rm.waiters[key], ch)
	rm.mu.Unlock()
	if !ok { // Start goroutine
		go func() {
			vx, err := f(key)
			rm.mu.Lock()
			r := resultHolder[K, V]{key: key, result: vx, err: err, resultTime: time.Now()}
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
				//time.Sleep(rm.deleteAfter)
				//rm.mu.Lock()
				//defer rm.mu.Unlock()
				//delete(rm.waiters, key)
				//delete(rm.response, key)
				rm.resultSliceMu.Lock()
				rm.resultSlice = append(rm.resultSlice, r)
				rm.resultSliceMu.Unlock()
			}()
		}()
	}
	res := <-ch
	return res.result, res.err
}

// Timeout returns a function that will call f with k and return the result or ErrTimeout if the duration is exceeded
func (rm *RequestMap[K, V]) Timeout(duration time.Duration, f func(k K) (V, error)) func(k K) (V, error) {
	return func(k K) (V, error) {
		res := make(chan resultHolder[K, V], 1)
		go func() {
			v, err := f(k)
			res <- resultHolder[K, V]{key: k, result: v, err: err}
		}()
		select {
		case r := <-res:
			return r.result, r.err
		case <-time.After(duration):
			var result V
			return result, ErrTimeout
		}
	}
}
