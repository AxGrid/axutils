package chans

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (19.03.2024)
*/

func TestShardChunkChan_ChunkAndShard(t *testing.T) {
	inc := make(chan int, 100)
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*100)
	shardChunk, err := NewShardChunk[int]().WithIncomingChan(inc).WithContext(ctx).WithShardCount(2).WithChunkSize(5).WithChunkTimeout(time.Millisecond * 50).WithShardFunc(func(m int) int {
		return m
	}).WithShardWorker(func(k int, m []int) {
		t.Logf("Worker:%d Recv: %v", k, m)
	}).Build()
	assert.Nil(t, err)
	assert.NotNil(t, shardChunk)

	// Вот это кидаем тестовые сообщения в канал
	for i := 0; i < 103; i++ {
		go func(k int) {
			inc <- k
		}(i)
	}
	time.Sleep(time.Millisecond * 200)
	cancelFn()
}
