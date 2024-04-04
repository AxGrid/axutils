package workers

import "context"

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (04.04.2024)
*/

type RunnerFn func()
type RunnerContextFn func(ctx context.Context)

type Runner struct {
	ctx        context.Context
	runners    chan RunnerFn
	runnersCtx chan RunnerContextFn
}

func (r *Runner) Run(fn RunnerFn) {
	r.runners <- fn
}

func (r *Runner) RunWithContext(fn RunnerContextFn) {
	r.runnersCtx <- fn
}

type RunnerBuilder struct {
	ctx         context.Context
	workerCount int
	bufferSize  int
}

func NewRunner() *RunnerBuilder {
	return &RunnerBuilder{
		ctx:         context.Background(),
		workerCount: 5,
		bufferSize:  100,
	}
}

func (b *RunnerBuilder) WithContext(ctx context.Context) *RunnerBuilder {
	b.ctx = ctx
	return b
}

func (b *RunnerBuilder) WithWorkerCount(workerCount int) *RunnerBuilder {
	b.workerCount = workerCount
	return b
}

func (b *RunnerBuilder) Build() *Runner {
	r := &Runner{
		ctx:        b.ctx,
		runners:    make(chan RunnerFn, b.bufferSize),
		runnersCtx: make(chan RunnerContextFn, b.bufferSize),
	}

	for i := 0; i < b.workerCount; i++ {
		go func() {
			for {
				select {
				case <-r.ctx.Done():
					return
				case fn := <-r.runners:
					fn()
				case fn := <-r.runnersCtx:
					fn(r.ctx)
				}
			}
		}()
	}
	return r
}
