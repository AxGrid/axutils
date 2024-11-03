package axmiddleware

import (
	"fmt"
	"time"
)

func CatchPanicMiddlewares[R, S any](ctx *Context[R, S]) {
	defer func() {
		if r := recover(); r != nil {
			ctx.logger.Error().Msgf("Panic: %v", r)
		}
	}()
	ctx.Next()
}

func LogRequestResponseMiddlewares[R, S any](ctx *Context[R, S]) {
	startTime := time.Now()
	ctx.WithValue("start_time", startTime)
	ctx.Next()
	ctx.logger.Debug().Interface("req", ctx.Request()).Interface("res", ctx.Response()).Str("total_time", fmt.Sprintf("%d ms.", time.Now().Sub(startTime).Milliseconds())).Msg("request")
}
