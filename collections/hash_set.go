package collections

import "sync"

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (16.05.2024)
*/

type HashSet[K comparable] struct {
	m        map[K]bool
	ml       []K
	mu       sync.RWMutex
	maxCount int
}

func NewHashSet[K comparable]() *HashSetBuilder[K] {
	return &HashSetBuilder[K]{
		maxCount: 0,
	}
}

func (s *HashSet[K]) Add(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.m[key] = true
	s.ml = append(s.ml, key)
	if s.maxCount > 0 && len(s.m) > s.maxCount {
		delete(s.m, s.ml[0])
		s.ml = s.ml[1:]
	}
}

func (s *HashSet[K]) Has(key K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.m[key]
	return ok
}

// HasOrCreate returns true if key already exists in the set, otherwise adds key to the set and returns false.
// If exist return true
func (s *HashSet[K]) HasOrCreate(key K) bool {
	s.mu.RLock()
	_, ok := s.m[key]
	s.mu.RUnlock()
	if ok {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok = s.m[key]
	if ok {
		return true
	}
	s.m[key] = true
	s.ml = append(s.ml, key)
	if s.maxCount > 0 && len(s.m) > s.maxCount {
		delete(s.m, s.ml[0])
		s.ml = s.ml[1:]
	}
	return false
}

func (s *HashSet[K]) Delete(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
	for i, k := range s.ml {
		if k == key {
			s.ml = append(s.ml[:i], s.ml[i+1:]...)
			break
		}
	}
}

func (s *HashSet[K]) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m)
}

func (s *HashSet[K]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m = make(map[K]bool)
	s.ml = make([]K, 0)
}

func (s *HashSet[K]) Keys() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]K, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys
}

type HashSetBuilder[K comparable] struct {
	maxCount int
}

func (b *HashSetBuilder[K]) WithMaxCount(maxCount int) *HashSetBuilder[K] {
	b.maxCount = maxCount
	return b
}

func (b *HashSetBuilder[K]) Build() *HashSet[K] {
	return &HashSet[K]{
		m:        make(map[K]bool),
		maxCount: b.maxCount,
	}
}
