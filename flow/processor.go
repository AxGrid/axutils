package flow

import "github.com/go-errors/errors"

type ProcessorBuilder[E, S any] struct {
	routes []ProcessorRouter[E, S]
}

func NewProcessor[E, S any]() *ProcessorBuilder[E, S] {
	return &ProcessorBuilder[E, S]{}
}

type ProcessorRouter[E, S any] struct {
	chain []func(event E, state S) error
}

type ProcessorRouterBuilder[E, S any] struct {
	parent *ProcessorBuilder[E, S]
	chain  []func(event E, state S) error
}

func (rb *ProcessorRouterBuilder[E, S]) On(cond func(event E, state S) bool) *ProcessorRouterBuilder[E, S] {
	rb.chain = append(rb.chain, func(event E, state S) error {
		if !cond(event, state) {
			return errors.New("filter not passed")
		}
		return nil
	})
	return rb
}

func (rb *ProcessorRouterBuilder[E, S]) OnEvent(cond func(e E) bool) *ProcessorRouterBuilder[E, S] {
	return rb.On(func(event E, _ S) bool {
		return cond(event)
	})
}

func (rb *ProcessorRouterBuilder[E, S]) Do(do func(event E, state S) error) *ProcessorRouterBuilder[E, S] {
	rb.chain = append(rb.chain, do)
	return rb
}

func (rb *ProcessorRouterBuilder[E, S]) build() {
	route := ProcessorRouter[E, S]{
		chain: rb.chain,
	}
	rb.parent.routes = append(rb.parent.routes, route)
}

func (fb *ProcessorBuilder[E, S]) Route(rb func(r *ProcessorRouterBuilder[E, S])) *ProcessorBuilder[E, S] {
	b := &ProcessorRouterBuilder[E, S]{
		parent: fb,
	}
	rb(b)
	b.build()
	return fb
}

func (fb *ProcessorBuilder[E, S]) RouteOn(cond func(event E) bool, do ...func(event E, state S) error) *ProcessorBuilder[E, S] {
	fb.Route(func(rb *ProcessorRouterBuilder[E, S]) {
		for _, fn := range do {
			rb.OnEvent(cond).Do(fn)
		}
	})
	return fb
}

func (fb *ProcessorBuilder[E, S]) Build() func(event E, state S) error {
	return func(event E, state S) error {
		for _, route := range fb.routes {
			for _, fn := range route.chain {
				if err := fn(event, state); err != nil {
					return err
				}
			}
		}
		return nil
	}
}
