package fms

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (09.04.2024)
*/

type FlowProcessorBuilder[E any, S any] struct{}

func NewFlow[E any, S any]() *FlowProcessorBuilder[E, S] {
	return &FlowProcessorBuilder[E, S]{}
}

type FlowProcessorRouterBuilder[E any, S any] struct {
	parent *FlowProcessorBuilder[E, S]
}

func (rb *FlowProcessorRouterBuilder[E, S]) On(cond func(event E, state S) bool) *FlowProcessorRouterBuilder[E, S] {
	return rb
}

func (rb *FlowProcessorRouterBuilder[E, S]) OnEvent(cond func(e E) bool) *FlowProcessorRouterBuilder[E, S] {
	return rb.On(func(event E, _ S) bool {
		return cond(event)
	})
}

func (rb *FlowProcessorRouterBuilder[E, S]) Do(do func(event E, state S) error) *FlowProcessorRouterBuilder[E, S] {
	return rb
}

func (fb *FlowProcessorBuilder[E, S]) Route(rb func(r *FlowProcessorRouterBuilder[E, S])) *FlowProcessorBuilder[E, S] {
	return fb
}

func (fb *FlowProcessorBuilder[E, S]) RouteOn(cond func(event E) bool, do ...func(event E, state S) error) *FlowProcessorBuilder[E, S] {
	fb.Route(func(rb *FlowProcessorRouterBuilder[E, S]) {
		rb.OnEvent(cond).Do(do)
	})
	return fb
}

func (fb *FlowProcessorBuilder[E, S]) Build() func(event E, state S) error {
	return func(event E, state S) error {
		return nil
	}
}
