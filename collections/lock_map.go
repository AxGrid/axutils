package collections

import "sync"

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (14.03.2024)
*/

type MapMutex[K comparable] struct {
	m  map[K]*sync.Mutex
	mu sync.RWMutex
}

func NewMapMutex[K comparable]() *MapMutex[K] {
	return &MapMutex[K]{
		m: make(map[K]*sync.Mutex),
	}
}

func (mm *MapMutex[K]) Lock(k K) {
	mm.mu.RLock()
	m, ok := mm.m[k]
	mm.mu.RUnlock()
	if !ok {
		mm.mu.Lock()
		m, ok = mm.m[k]
		if !ok {
			m = &sync.Mutex{}
			mm.m[k] = m
		}
		mm.mu.Unlock()
	}
	m.Lock()
}

func (mm *MapMutex[K]) Unlock(k K) {
	mm.mu.RLock()
	m, ok := mm.m[k]
	mm.mu.RUnlock()
	if ok {
		m.Unlock()
	}
}

type MapRWMutex[K comparable] struct {
	m  map[K]*sync.RWMutex
	mu sync.RWMutex
}

func NewMapRWMutex[K comparable]() *MapRWMutex[K] {
	return &MapRWMutex[K]{
		m: make(map[K]*sync.RWMutex),
	}
}

func (mm *MapRWMutex[K]) RLock(k K) {
	mm.mu.RLock()
	m, ok := mm.m[k]
	mm.mu.RUnlock()
	if !ok {
		mm.mu.Lock()
		m, ok = mm.m[k]
		if !ok {
			m = &sync.RWMutex{}
			mm.m[k] = m
		}
		mm.mu.Unlock()
	}
	m.RLock()
}

func (mm *MapRWMutex[K]) RUnlock(k K) {
	mm.mu.RLock()
	m, ok := mm.m[k]
	mm.mu.RUnlock()
	if ok {
		m.RUnlock()
	}
}

func (mm *MapRWMutex[K]) Lock(k K) {
	mm.mu.RLock()
	m, ok := mm.m[k]
	mm.mu.RUnlock()
	if !ok {
		mm.mu.Lock()
		m, ok = mm.m[k]
		if !ok {
			m = &sync.RWMutex{}
			mm.m[k] = m
		}
		mm.mu.Unlock()
	}
	m.Lock()
}

func (mm *MapRWMutex[K]) Unlock(k K) {
	mm.mu.RLock()
	m, ok := mm.m[k]
	mm.mu.RUnlock()
	if ok {
		m.Unlock()
	}
}
