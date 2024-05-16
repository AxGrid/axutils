package flow

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (05.05.2024)
*/

type Context[S comparable] interface {
	GetState() S
}

type Flow[E any, S comparable, T Context[S]] struct {
}

func (f *Flow[E, S, T]) Process(state T, event E) error {
	return nil
}

type Builder[E any, S comparable, T Context[S]] struct {
	states map[S]StateRouter[E, S, T]
}

func NewFlow[E any, S comparable, T Context[S]]() *Builder[E, S, T] {
	return &Builder[E, S, T]{}
}

func (b *Builder[E, S, T]) Route(route func(r *StateRouter[E, S, T])) *Builder[E, S, T] {
	return b
}

type StateRouter[E any, S comparable, T Context[S]] struct {
	chain    []func(event E, state S) error
	errChain []func(event E, state S, err error) error
}
