package collections

import "sync"

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (13.05.2024)
*/

type TwoKeyMap[K1, K2 comparable, V any] struct {
	data   map[K1]V
	mu     sync.RWMutex
	k1Tok2 map[K1]K2
	k2Tok1 map[K2]K1
}

func NewTwoKeyMap[K1, K2 comparable, V any]() *TwoKeyMap[K1, K2, V] {
	return &TwoKeyMap[K1, K2, V]{
		data:   make(map[K1]V),
		k1Tok2: make(map[K1]K2),
		k2Tok1: make(map[K2]K1),
	}
}

func (m *TwoKeyMap[K1, K2, V]) Put(k1 K1, k2 K2, v V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[k1] = v
	m.k1Tok2[k1] = k2
	m.k2Tok1[k2] = k1
}

func (m *TwoKeyMap[K1, K2, V]) Get(k1 K1) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[k1]
	return v, ok
}

func (m *TwoKeyMap[K1, K2, V]) GetByK2(k2 K2) (v V, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	k1, ok := m.k2Tok1[k2]
	if !ok {
		ok = false
		return
	}
	v, ok = m.data[k1]
	return v, ok
}

func (m *TwoKeyMap[K1, K2, V]) Remove(k1 K1) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, k1)
	delete(m.k2Tok1, m.k1Tok2[k1])
	delete(m.k1Tok2, k1)
}

func (m *TwoKeyMap[K1, K2, V]) RemoveByK2(k2 K2) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, m.k2Tok1[k2])
	delete(m.k1Tok2, m.k2Tok1[k2])
	delete(m.k2Tok1, k2)
}

func (m *TwoKeyMap[K1, K2, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

func (m *TwoKeyMap[K1, K2, V]) Keys() []K1 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]K1, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *TwoKeyMap[K1, K2, V]) KeysK2() []K2 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]K2, 0, len(m.k2Tok1))
	for k := range m.k2Tok1 {
		keys = append(keys, k)
	}
	return keys
}
