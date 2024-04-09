package flow2

import "github.com/go-errors/errors"

type FlowProcessorBuilder[E, S any] struct {
	routes []FlowProcessorRouter[E, S]
}

func NewFlow[E, S any]() *FlowProcessorBuilder[E, S] {
	return &FlowProcessorBuilder[E, S]{}
}

type FlowProcessorRouter[E, S any] struct {
	chain []func(event E, state S) error
}

type FlowProcessorRouterBuilder[E, S any] struct {
	parent *FlowProcessorBuilder[E, S]
	chain  []func(event E, state S) error
}

func (rb *FlowProcessorRouterBuilder[E, S]) On(cond func(event E, state S) bool) *FlowProcessorRouterBuilder[E, S] {
	rb.chain = append(rb.chain, func(event E, state S) error {
		if !cond(event, state) {
			return errors.New("filter not passed")
		}
		return nil
	})
	return rb
}

func (rb *FlowProcessorRouterBuilder[E, S]) OnEvent(cond func(e E) bool) *FlowProcessorRouterBuilder[E, S] {
	return rb.On(func(event E, _ S) bool {
		return cond(event)
	})
}

func (rb *FlowProcessorRouterBuilder[E, S]) Do(do func(event E, state S) error) *FlowProcessorRouterBuilder[E, S] {
	rb.chain = append(rb.chain, do)
	return rb
}

func (rb *FlowProcessorRouterBuilder[E, S]) Build() {
	route := FlowProcessorRouter[E, S]{
		chain: rb.chain,
	}
	rb.parent.routes = append(rb.parent.routes, route)
}

func (fb *FlowProcessorBuilder[E, S]) Route(rb func(r *FlowProcessorRouterBuilder[E, S])) *FlowProcessorBuilder[E, S] {
	b := &FlowProcessorRouterBuilder[E, S]{
		parent: fb,
	}
	rb(b)
	return fb
}

func (fb *FlowProcessorBuilder[E, S]) RouteOn(cond func(event E) bool, do ...func(event E, state S) error) *FlowProcessorBuilder[E, S] {
	fb.Route(func(rb *FlowProcessorRouterBuilder[E, S]) {
		for _, fn := range do {
			rb.OnEvent(cond).Do(fn).Build()
		}
	})
	return fb
}

func (fb *FlowProcessorBuilder[E, S]) Build() func(event E, state S) error {
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
