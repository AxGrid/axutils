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

func (s *Set[T]) Remove(v T) {
	delete(s.m, v)
}

func (s *Set[T]) Has(v T) bool {
	_, ok := s.m[v]
	return ok
}

func (s *Set[T]) Values() []T {
	values := make([]T, 0, len(s.m))
	for v := range s.m {
		values = append(values, v)
	}
	return values
}

func (s *Set[T]) Size() int {
	return len(s.m)
}

func (s *Set[T]) Clear() {
	clear(s.m)
}
