package collections

import (
	"sync"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (14.03.2024)
*/

type GuavaLoadFunc[K comparable, V any] func(K) (V, error)
type GuavaMap[K comparable, V any] struct {
	stored      map[K]V
	storedSlice []K
	mu          sync.RWMutex
	maxCount    int
	loadFunc    GuavaLoadFunc[K, V]
}

func (m *GuavaMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.stored, key)
	if m.maxCount > 0 {
		for i, k := range m.storedSlice {
			if k == key {
				m.storedSlice = append(m.storedSlice[:i], m.storedSlice[i+1:]...)
				break
			}
		}
	}
}

func (m *GuavaMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stored = make(map[K]V)
	m.storedSlice = make([]K, 0, m.maxCount)
}

func (m *GuavaMap[K, V]) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.stored)
}

func (m *GuavaMap[K, V]) Get(key K) (V, error) {
	m.mu.RLock()
	val, ok := m.stored[key]
	m.mu.RUnlock()
	if ok {
		return val, nil
	}
	if m.loadFunc == nil {
		return val, nil
	}
	loadedVal, err := m.loadFunc(key)
	if err != nil {
		return val, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok = m.stored[key]
	if ok {
		return val, nil
	}
	if m.maxCount > 0 {
		if len(m.storedSlice) >= m.maxCount {
			delete(m.stored, m.storedSlice[0])
			m.storedSlice = m.storedSlice[1:]
		}
		m.storedSlice = append(m.storedSlice, key)
	}
	m.stored[key] = loadedVal
	return loadedVal, nil
}

func (m *GuavaMap[K, V]) Set(key K, val V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.stored[key]
	m.stored[key] = val
	if ok {
		return
	}
	if m.maxCount > 0 {
		if len(m.storedSlice) >= m.maxCount {
			delete(m.stored, m.storedSlice[0])
			m.storedSlice = m.storedSlice[1:]
		}
		m.storedSlice = append(m.storedSlice, key)
	}
}

type GuavaMapBuilder[K comparable, V any] struct {
	loadFunc GuavaLoadFunc[K, V]
	maxCount int
}

func NewGuavaMap[K comparable, V any]() *GuavaMapBuilder[K, V] {
	return &GuavaMapBuilder[K, V]{}
}

func (b *GuavaMapBuilder[K, V]) WithLoadFunc(loadFunc GuavaLoadFunc[K, V]) *GuavaMapBuilder[K, V] {
	b.loadFunc = loadFunc
	return b
}

func (b *GuavaMapBuilder[K, V]) WithMaxCount(maxCount int) *GuavaMapBuilder[K, V] {
	b.maxCount = maxCount
	return b
}

func (b *GuavaMapBuilder[K, V]) Build() *GuavaMap[K, V] {
	res := &GuavaMap[K, V]{
		stored:   make(map[K]V),
		loadFunc: b.loadFunc,
		maxCount: b.maxCount,
	}
	if b.maxCount > 0 {
		res.storedSlice = make([]K, 0, b.maxCount)
	}
	return res
}
