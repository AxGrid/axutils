package axmiddleware

import (
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
zed (13.02.2024)
*/

type Request struct {
	Value int
}

type Response struct {
	Result int
}

func SumMiddleware(ctx *Context[*Request, *Response]) {
	ctx.Response().Result = ctx.Request().Value * 10
	ctx.Next()
}

func LogMiddleware(ctx *Context[*Request, *Response]) {
	ctx.Next()
	ctx.Logger().Error().Interface("req", ctx.Request()).Interface("res", ctx.Response()).Msg("logger")
}

func TestNewContextProcessorBuilder(t *testing.T) {
	proc := NewProcessor[*Request, *Response]().WithLogger(log.Logger).WithMiddlewares(SumMiddleware, LogMiddleware).WithHandlers(func(ctx *Context[*Request, *Response]) {
		log.Debug().Msgf("result: %d", ctx.Response().Result)
	}).Build()
	if proc == nil {
		t.Fatal("builder is nil")
	}
	var response Response
	proc.Process(nil, &Request{Value: 10}, &response)
	assert.Equal(t, 100, response.Result)
}
