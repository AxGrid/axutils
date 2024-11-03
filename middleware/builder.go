package axmiddleware

import (
	"context"
	"github.com/rs/zerolog"
)

/*
zed (13.02.2024)
*/

type ContextProcessorBuilder[R, S any] struct {
	logger      zerolog.Logger
	ctx         context.Context
	middlewares HandlersChain[R, S]
	handlers    HandlersChain[R, S]
}

func NewProcessor[R, S any]() *ContextProcessorBuilder[R, S] {
	return &ContextProcessorBuilder[R, S]{
		logger: zerolog.Nop(),
		ctx:    context.Background(),
	}
}

func (b *ContextProcessorBuilder[R, S]) WithLogger(logger zerolog.Logger) *ContextProcessorBuilder[R, S] {
	b.logger = logger
	return b
}

func (b *ContextProcessorBuilder[R, S]) WithContext(ctx context.Context) *ContextProcessorBuilder[R, S] {
	b.ctx = ctx
	return b
}

func (b *ContextProcessorBuilder[R, S]) WithMiddlewares(middlewares ...HandlerFunc[R, S]) *ContextProcessorBuilder[R, S] {
	b.middlewares = append(b.middlewares, middlewares...)
	return b
}

func (b *ContextProcessorBuilder[R, S]) WithHandlers(handlers ...HandlerFunc[R, S]) *ContextProcessorBuilder[R, S] {
	b.handlers = append(b.handlers, handlers...)
	return b
}

func (b *ContextProcessorBuilder[R, S]) Build() MiddlewareProcessor[R, S] {
	chain := make(HandlersChain[R, S], 0, len(b.middlewares)+1)
	for _, middleware := range b.middlewares {
		chain = append(chain, middleware)
	}
	chain = append(chain, func(c *Context[R, S]) {
		for _, handler := range b.handlers {
			handler(c)
		}
	})
	res := &middlewareProcessor[R, S]{
		logger:        b.logger,
		ctx:           b.ctx,
		handlersChain: chain,
	}
	return res
}
