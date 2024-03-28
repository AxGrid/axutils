package chans

import (
	"context"
	"github.com/go-errors/errors"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (19.03.2024)
*/

var ErrShardFuncIsNil = errors.New("shard func is nil")

type ShardFunc[T any] func(T) int

type ShardChan[T any] struct {
	ctx           context.Context
	shardFunc     ShardFunc[T]
	incomingChan  chan T
	shardCount    int
	outgoingChans map[int]chan T
}

func (s *ShardChan[T]) ShardCount() int {
	return s.shardCount
}

func (s *ShardChan[T]) C(key int) chan T {
	if key < 0 || key >= s.shardCount {
		return nil
	}
	return s.outgoingChans[key]
}

func (s *ShardChan[T]) Add(msg T) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	case s.incomingChan <- msg:
		return nil
	}
}

func (s *ShardChan[T]) Sizes() []int {
	sizes := make([]int, s.shardCount)
	for i := 0; i < s.shardCount; i++ {
		sizes[i] = len(s.outgoingChans[i])
	}
	return sizes
}

type ShardChanBuilder[T any] struct {
	ctx                context.Context
	shardFunc          ShardFunc[T]
	shardCount         int
	incomingBufferSize int
	outgoingBufferSize int
	incomingChan       chan T
}

func NewShardChan[T any]() *ShardChanBuilder[T] {
	return &ShardChanBuilder[T]{
		incomingBufferSize: 1000,
		outgoingBufferSize: 100,
		shardCount:         4,
		ctx:                context.Background(),
	}
}

func (b *ShardChanBuilder[T]) WithContext(ctx context.Context) *ShardChanBuilder[T] {
	b.ctx = ctx
	return b
}

func (b *ShardChanBuilder[T]) WithShardFunc(shardFunc ShardFunc[T]) *ShardChanBuilder[T] {
	b.shardFunc = shardFunc
	return b
}

func (b *ShardChanBuilder[T]) WithShardCount(shardCount int) *ShardChanBuilder[T] {
	b.shardCount = shardCount
	return b
}

func (b *ShardChanBuilder[T]) WithIncomingBufferSize(incomingBufferSize int) *ShardChanBuilder[T] {
	b.incomingBufferSize = incomingBufferSize
	return b
}

func (b *ShardChanBuilder[T]) WithOutgoingBufferSize(outgoingBufferSize int) *ShardChanBuilder[T] {
	b.outgoingBufferSize = outgoingBufferSize
	return b
}

func (b *ShardChanBuilder[T]) WithIncomingChan(incomingChan chan T) *ShardChanBuilder[T] {
	b.incomingChan = incomingChan
	return b
}

func (b *ShardChanBuilder[T]) Build() (*ShardChan[T], error) {
	if b.shardFunc == nil {
		return nil, ErrShardFuncIsNil
	}
	if b.incomingChan == nil {
		b.incomingChan = make(chan T, b.incomingBufferSize)
	}
	res := &ShardChan[T]{
		ctx:           b.ctx,
		shardFunc:     b.shardFunc,
		incomingChan:  b.incomingChan,
		shardCount:    b.shardCount,
		outgoingChans: make(map[int]chan T, b.shardCount),
	}
	for i := 0; i < b.shardCount; i++ {
		res.outgoingChans[i] = make(chan T, b.outgoingBufferSize)
	}
	go func() {
		for {
			select {
			case <-res.ctx.Done():
				return
			case msg := <-res.incomingChan:
				key := res.shardFunc(msg) % res.shardCount
				res.outgoingChans[key] <- msg
			}
		}
	}()
	return res, nil
}
