package collections

import "sync"

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (11.04.2024)
*/

type Map interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Has(key string) bool
	SetIfNotExists(key string, value interface{}) bool
	Size() int
}

type SimpleMap[K comparable, V any] struct {
	m  map[K]V
	mu sync.RWMutex
}

func NewSimpleMap[K comparable, V any]() *SimpleMap[K, V] {
	return &SimpleMap[K, V]{
		m: make(map[K]V),
	}
}

func (m *SimpleMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.m[key]
	return v, ok
}

func (m *SimpleMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = value
}

func (m *SimpleMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, key)
}

func (m *SimpleMap[K, V]) Has(key K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.m[key]
	return ok
}

func (m *SimpleMap[K, V]) SetIfNotExists(key K, value V) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.m[key]; ok {
		return false
	}
	m.m[key] = value
	return true
}

func (m *SimpleMap[K, V]) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.m)
}
