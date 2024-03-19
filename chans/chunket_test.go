package chans

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
zed (19.03.2024)
*/

func TestChunker_Chunk(t *testing.T) {
	chunker := NewChunkChan[int]().WithChunkSize(5).WithChunkTimeout(time.Millisecond * 100).Build()
	assert.NotNil(t, chunker)
	for i := 0; i < 100; i++ {
		chunker.Add(i)
	}
	for {
		chunk := <-chunker.C()
		t.Logf("Recv: %v", chunk)
		if len(chunk) == 0 || chunk[len(chunk)-1] == 99 {
			break
		}
	}
}

func TestChunker_ChunkTimeout(t *testing.T) {
	chunker := NewChunkChan[int]().WithChunkSize(5).WithChunkTimeout(time.Millisecond * 100).Build()
	assert.NotNil(t, chunker)
	for i := 0; i < 4; i++ {
		chunker.Add(i)
	}
	chunk := <-chunker.C()
	t.Logf("Recv: %v", chunk)
}

func TestChunker_ChunkAndShard(t *testing.T) {
	inc := make(chan *simpleAxChunkAndShardMessage, 100)
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*100)
	sharder, err := NewShardChan[*simpleAxChunkAndShardMessage]().WithContext(ctx).WithShardCount(2).WithIncomingChan(inc).WithShardFunc(func(m *simpleAxChunkAndShardMessage) int {
		return m.User
	}).Build()
	assert.Nil(t, err)
	for i := 0; i < sharder.ShardCount(); i++ {
		chunker := NewChunkChan[*simpleAxChunkAndShardMessage]().WithChunkSize(5).WithChunkTimeout(time.Millisecond * 50).WithContext(ctx).Build()
		go func(k int) {
			for {
				select {
				case <-ctx.Done():
					return
				case m := <-sharder.C(k):
					chunker.Add(m)
				}
			}
		}(i)
		go func(k int) {
			for {
				select {
				case <-ctx.Done():
					return
				case batch := <-chunker.C():
					var users []int
					for _, msg := range batch {
						users = append(users, msg.User)
					}
					t.Logf("Shard:%d Recv: %v", k, users)
					for _, msg := range batch {
						msg.Ack <- nil
					}
				}
			}
		}(i)
	}
	// Вот тут уже сыпим собщения в канал
	processed := int32(0)
	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(k int) {
			m := &simpleAxChunkAndShardMessage{User: k, Ack: make(chan error)}
			inc <- m
			<-m.Ack
			wg.Done()
			atomic.AddInt32(&processed, 1)
		}(i)
	}
	wg.Wait()
	assert.Equal(t, int32(100), processed)
	cancelFn()
}

type simpleAxChunkAndShardMessage struct {
	User int
	Ack  chan error
}

func (s *simpleAxChunkAndShardMessage) GetShardKey() int {
	return s.User % 2
}
