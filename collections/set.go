package collections

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (04.05.2024)
*/

type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable](init ...T) *Set[T] {
	t := &Set[T]{m: make(map[T]struct{})}
	for _, v := range init {
		t.Add(v)
	}
	return t
}

func (s *Set[T]) Add(v T) {
	s.m[v] = struct{}{}
}

func (s *Set[T]) AddAll(v ...T) {
	for _, i := range v {
		s.Add(i)
	}
}

func (s *Set[T]) Remove(v T) {
	delete(s.m, v)
}

func (s *Set[T]) Has(v T) bool {
	_, ok := s.m[v]
	return ok
}

func (s *Set[T]) HasAll(v ...T) bool {
	for _, i := range v {
		if !s.Has(i) {
			return false
		}
	}
	return true
}

func (s *Set[T]) HasAny(v ...T) bool {
	for _, i := range v {
		if s.Has(i) {
			return true
		}
	}
	return false
}

func (s *Set[T]) Values() []T {
	values := make([]T, 0, len(s.m))
	for v := range s.m {
		values = append(values, v)
	}
	return values
}

func (s *Set[T]) All() []T {
	return s.Values()
}

func (s *Set[T]) Equals(other *Set[T]) bool {
	if s.Size() != other.Size() {
		return false
	}
	for _, v := range s.Values() {
		if !other.Has(v) {
			return false
		}
	}
	return true
}

func (s *Set[T]) Clone() *Set[T] {
	clone := NewSet[T]()
	for _, v := range s.Values() {
		clone.Add(v)
	}
	return clone
}

func (s *Set[T]) Size() int {
	return len(s.m)
}

func (s *Set[T]) Clear() {
	clear(s.m)
}
