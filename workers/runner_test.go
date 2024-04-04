package workers

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (04.04.2024)
*/

func TestNewRunner(t *testing.T) {
	r := NewRunner().
		WithContext(context.Background()).
		WithWorkerCount(3).
		Build()
	wg := sync.WaitGroup{}
	done := int32(0)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		r.Run(func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&done, 1)
		})
	}
	wg.Wait()
	assert.Equal(t, int32(10), done)
}
