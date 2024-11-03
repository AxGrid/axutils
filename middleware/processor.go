package axmiddleware

import (
	"context"
	"github.com/rs/zerolog"
)

type MiddlewareProcessor[R, S any] interface {
	Process(ctx context.Context, request R, response S) (int, error)
}

type middlewareProcessor[R, S any] struct {
	logger        zerolog.Logger
	ctx           context.Context
	handlersChain HandlersChain[R, S]
}

// Process is a method of middlewareProcessor that processes the request and response
// returns the status code and error
func (b *middlewareProcessor[R, S]) Process(ctx context.Context, request R, response S) (int, error) {
	if ctx == nil {
		ctx = b.ctx
	}
	rctx := newContext(ctx, request, response)
	rctx.logger = &b.logger
	rctx.handlers = b.handlersChain
	rctx.Next()

	return rctx.statusCode, rctx.error
}
