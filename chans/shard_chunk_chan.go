package chans

import (
	"context"
	"time"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (19.03.2024)
*/

type ShardChunk[T any] struct {
	sharder  *ShardChan[T]
	chunkers []*ChunkChan[T]
}

type ShardChunkFunc[T any] func(int, []T)

func (s *ShardChunk[T]) ShardCount() int {
	return s.sharder.ShardCount()
}

func (s *ShardChunk[T]) C(key int) chan []T {
	if key < 0 || key >= s.sharder.ShardCount() {
		return nil
	}
	return s.chunkers[key].C()
}

type ShardChunkBuilder[T any] struct {
	ctx                context.Context
	chunkSize          int
	chunkTimeout       time.Duration
	shardCount         int
	shardFunc          ShardFunc[T]
	shardWorker        ShardChunkFunc[T]
	incomingChan       chan T
	incomingBufferSize int
	outgoingBufferSize int
}

func NewShardChunk[T any]() *ShardChunkBuilder[T] {
	return &ShardChunkBuilder[T]{
		ctx:                context.Background(),
		incomingBufferSize: 1000,
		outgoingBufferSize: 100,
		shardCount:         4,
		chunkTimeout:       time.Millisecond * 50,
		chunkSize:          500,
	}
}

func (b *ShardChunkBuilder[T]) WithShardWorker(shardWorker ShardChunkFunc[T]) *ShardChunkBuilder[T] {
	b.shardWorker = shardWorker
	return b
}

func (b *ShardChunkBuilder[T]) WithContext(ctx context.Context) *ShardChunkBuilder[T] {
	b.ctx = ctx
	return b
}

func (b *ShardChunkBuilder[T]) WithChunkSize(chunkSize int) *ShardChunkBuilder[T] {
	b.chunkSize = chunkSize
	return b
}

func (b *ShardChunkBuilder[T]) WithChunkTimeout(chunkTimeout time.Duration) *ShardChunkBuilder[T] {
	b.chunkTimeout = chunkTimeout
	return b
}

func (b *ShardChunkBuilder[T]) WithShardCount(shardCount int) *ShardChunkBuilder[T] {
	b.shardCount = shardCount
	return b
}

func (b *ShardChunkBuilder[T]) WithShardFunc(shardFunc ShardFunc[T]) *ShardChunkBuilder[T] {
	b.shardFunc = shardFunc
	return b
}

func (b *ShardChunkBuilder[T]) WithIncomingBufferSize(incomingBufferSize int) *ShardChunkBuilder[T] {
	b.incomingBufferSize = incomingBufferSize
	return b
}

func (b *ShardChunkBuilder[T]) WithOutgoingBufferSize(outgoingBufferSize int) *ShardChunkBuilder[T] {
	b.outgoingBufferSize = outgoingBufferSize
	return b
}

func (b *ShardChunkBuilder[T]) WithIncomingChan(incomingChan chan T) *ShardChunkBuilder[T] {
	b.incomingChan = incomingChan
	return b
}

func (b *ShardChunkBuilder[T]) Build() (*ShardChunk[T], error) {
	if b.incomingChan == nil {
		b.incomingChan = make(chan T, b.incomingBufferSize)
	}
	sharder, err := NewShardChan[T]().WithContext(b.ctx).WithOutgoingBufferSize(b.incomingBufferSize / b.shardCount).WithShardCount(b.shardCount).WithShardFunc(b.shardFunc).WithIncomingChan(b.incomingChan).Build()
	if err != nil {
		return nil, err
	}
	chunkers := make([]*ChunkChan[T], b.shardCount)
	for i := 0; i < b.shardCount; i++ {
		chunkers[i] = NewChunkChan[T]().WithContext(b.ctx).WithChunkSize(b.chunkSize).WithOutgoingBufferSize(b.outgoingBufferSize).WithChunkTimeout(b.chunkTimeout).Build()
		go func(k int) {
			for {
				select {
				case <-b.ctx.Done():
					return
				case m := <-sharder.C(k):
					chunkers[k].Add(m)
				}
			}
		}(i)
	}
	if b.shardWorker != nil {
		for i := 0; i < b.shardCount; i++ {
			go func(k int) {
				for {
					select {
					case <-b.ctx.Done():
						return
					case batch := <-chunkers[k].C():
						b.shardWorker(k, batch)
					}
				}
			}(i)
		}
	}

	return &ShardChunk[T]{
		sharder:  sharder,
		chunkers: chunkers,
	}, nil
}
