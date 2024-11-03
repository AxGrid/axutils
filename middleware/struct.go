package axmiddleware

import (
	"context"
	"github.com/rs/zerolog"
	"math"
)

/*
zed (13.02.2024)
*/

type HandlerFunc[R, S any] func(*Context[R, S])
type HandlersChain[R, S any] []HandlerFunc[R, S]

const abortIndex int8 = math.MaxInt8 >> 1

type Context[R, S any] struct {
	ctx        context.Context
	logger     *zerolog.Logger
	request    R
	response   S
	handlers   HandlersChain[R, S]
	index      int8
	error      error
	statusCode int
}

func (c *Context[R, S]) Logger() *zerolog.Logger {
	return c.logger
}

func (c *Context[R, Q]) Request() R {
	return c.request
}

func (c *Context[P, S]) Response() S {
	return c.response
}

func (c *Context[R, S]) GetError() error {
	return c.error
}

func (c *Context[R, S]) Error(err error) {
	c.error = err
}

func (c *Context[R, S]) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context[R, S]) Abort() {
	c.index = abortIndex
}

func (c *Context[R, S]) IsAborted() bool {
	return c.index >= abortIndex
}

func (c *Context[R, S]) StatusCode() int {
	return c.statusCode
}

func (c *Context[R, S]) Context() context.Context {
	return c.ctx
}

func (c *Context[R, S]) AbortWithError(err error) {
	c.Error(err)
	c.Abort()
	c.statusCode = 500
}
func (c *Context[R, S]) AbortWithErrorAndCode(code int, err error) {
	c.Error(err)
	c.Abort()
	c.statusCode = code
}

func newContext[R, S any](ctx context.Context, request R, response S) *Context[R, S] {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Context[R, S]{
		ctx:      ctx,
		index:    -1,
		request:  request,
		response: response,
	}
}

func (c *Context[R, S]) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c *Context[R, S]) WithValue(key, val interface{}) *Context[R, S] {
	c.ctx = context.WithValue(c.ctx, key, val)
	return c
}

func (c *Context[R, S]) Set(key, val interface{}) *Context[R, S] {
	return c.WithValue(key, val)
}

func (c *Context[R, S]) MustValue(key interface{}) interface{} {
	val := c.Value(key)
	if val == nil {
		panic("value not found")
	}
	return val
}

func (c *Context[R, S]) MustStringValue(key interface{}) string {
	val := c.MustValue(key)
	str, ok := val.(string)
	if !ok {
		panic("value is not a string")
	}
	return str
}

func (c *Context[R, S]) MustIntValue(key interface{}) int {
	val := c.MustValue(key)
	i, ok := val.(int)
	if !ok {
		panic("value is not an int")
	}
	return i
}

func (c *Context[R, S]) MustInt64Value(key interface{}) int64 {
	val := c.MustValue(key)
	i, ok := val.(int64)
	if !ok {
		panic("value is not an int64")
	}
	return i
}

func (c *Context[R, S]) MustUint64Value(key interface{}) uint64 {
	val := c.MustValue(key)
	i, ok := val.(uint64)
	if !ok {
		panic("value is not an int64")
	}
	return i
}

func (c *Context[R, S]) MustFloat64Value(key interface{}) float64 {
	val := c.MustValue(key)
	i, ok := val.(float64)
	if !ok {
		panic("value is not a float64")
	}
	return i
}

func (c *Context[R, S]) MustBoolValue(key interface{}) bool {
	val := c.MustValue(key)
	i, ok := val.(bool)
	if !ok {
		panic("value is not a bool")
	}
	return i
}

func (c *Context[R, S]) MustStringValueSlice(key interface{}) []string {
	val := c.MustValue(key)
	i, ok := val.([]string)
	if !ok {
		panic("value is not a []string")
	}
	return i
}
