package collections

import (
	"context"
	"github.com/go-errors/errors"
	"sync"
	"time"
)

type holderMapHolder[V any] struct {
	Object          V
	Err             error
	inChan          chan error
	released        bool
	releaseChanList []chan error
	timeOutTimer    *time.Timer
	mu              sync.Mutex
}
type HolderMap[K comparable, V any] struct {
	mu                     sync.RWMutex
	m                      map[K]*holderMapHolder[V]
	ctx                    context.Context
	timeoutDuration        time.Duration
	destroyElementDuration time.Duration
}

func (c *HolderMap[K, V]) Get(trx K) (V, error) {
	c.mu.RLock()
	h, ok := c.m[trx]
	c.mu.RUnlock()
	if ok {
		return h.Object, h.Err
	}
	var defaultV V
	return defaultV, errors.New("not found")
}

func (c *HolderMap[K, V]) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.m)
}

func (c *HolderMap[K, V]) Wait(trx K, target V) error {
	c.mu.RLock()
	h, ok := c.m[trx]
	c.mu.RUnlock()
	if ok { // Has holder
		if h.released {
			return h.Err
		} else {
			waitChan := make(chan error, 1)
			h.mu.Lock()
			h.releaseChanList = append(h.releaseChanList, waitChan)
			h.mu.Unlock()
			return <-waitChan
		}
	} else {
		c.mu.Lock()
		h, ok = c.m[trx]
		if ok {
			c.mu.Unlock()
			if h.released {
				return h.Err
			} else {
				waitChan := make(chan error, 1)
				h.mu.Lock()
				h.releaseChanList = append(h.releaseChanList, waitChan)
				h.mu.Unlock()
				return <-waitChan
			}
		} else {
			waitChan := make(chan error, 1)
			h = &holderMapHolder[V]{
				Object:          target,
				inChan:          make(chan error, 1),
				releaseChanList: []chan error{waitChan},
				timeOutTimer:    time.NewTimer(c.timeoutDuration),
			}
			go func() {
				defer h.timeOutTimer.Stop()
				select {
				case <-h.timeOutTimer.C:
					h.mu.Lock()
					h.released = true
					h.Err = errors.New("timeout")
				case <-c.ctx.Done():
					return
				case errInChan := <-h.inChan:
					h.mu.Lock()
					h.released = true
					h.Err = errInChan
				}
				for _, ch := range h.releaseChanList {
					ch <- h.Err
				}
				h.mu.Unlock()
				go func() {
					time.Sleep(c.destroyElementDuration)
					c.mu.Lock()
					delete(c.m, trx)
					c.mu.Unlock()
				}()
			}()
			c.m[trx] = h
			c.mu.Unlock()
			return <-waitChan
		}
	}
}

func (c *HolderMap[K, V]) Update(trx K, target V) error {
	c.mu.RLock()
	h, ok := c.m[trx]
	c.mu.RUnlock()
	if ok {
		h.mu.Lock()
		h.Object = target
		h.mu.Unlock()
		return nil
	} else {
		return errors.New("not found")
	}
}

func (c *HolderMap[K, V]) Release(trx K) error {
	c.mu.RLock()
	h, ok := c.m[trx]
	c.mu.RUnlock()
	if ok {
		h.inChan <- nil
		return nil
	} else {
		return errors.New("not found")
	}
}

func (c *HolderMap[K, V]) Error(trx K, err error) error {
	c.mu.RLock()
	h, ok := c.m[trx]
	c.mu.RUnlock()
	if ok {
		h.inChan <- err
		return nil
	} else {
		return errors.New("not found")
	}
}

type HolderMapBuilder[K comparable, V any] struct {
	ctx                    context.Context
	timeoutDuration        time.Duration
	destroyElementDuration time.Duration
}

func NewHolderMap[K comparable, V any]() *HolderMapBuilder[K, V] {
	return &HolderMapBuilder[K, V]{
		ctx:                    context.Background(),
		timeoutDuration:        time.Second * 10,
		destroyElementDuration: time.Minute * 5,
	}
}

func (c *HolderMapBuilder[K, V]) WithContext(ctx context.Context) *HolderMapBuilder[K, V] {
	c.ctx = ctx
	return c
}

func (c *HolderMapBuilder[K, V]) WithTimeout(d time.Duration) *HolderMapBuilder[K, V] {
	c.timeoutDuration = d
	return c
}

func (c *HolderMapBuilder[K, V]) WithTTL(d time.Duration) *HolderMapBuilder[K, V] {
	c.destroyElementDuration = d
	return c
}

func (c *HolderMapBuilder[K, V]) Build() *HolderMap[K, V] {
	return &HolderMap[K, V]{
		m:                      make(map[K]*holderMapHolder[V]),
		ctx:                    c.ctx,
		timeoutDuration:        c.timeoutDuration,
		destroyElementDuration: c.destroyElementDuration,
	}
}
