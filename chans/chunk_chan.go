package chans

import (
	"context"
	"time"
)

type ChunkFunc[T any] func([]T)

type ChunkChan[T any] struct {
	ctx          context.Context
	chunkSize    int
	chunkTimeout time.Duration
	incomingChan chan T
	outgoingChan chan []T
}

func (c *ChunkChan[T]) Add(item T) {
	c.incomingChan <- item
}

func (c *ChunkChan[T]) Incoming() chan T {
	return c.incomingChan
}

func (c *ChunkChan[T]) C() chan []T {
	return c.outgoingChan
}

func (c *ChunkChan[T]) Sizes() (int, int) {
	return len(c.incomingChan), len(c.outgoingChan)
}

func (c *ChunkChan[T]) run() {
	t := time.NewTicker(c.chunkTimeout)
	defer t.Stop()
	chunk := make([]T, 0, c.chunkSize)
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-t.C:
			if len(chunk) > 0 {
				c.outgoingChan <- chunk
				chunk = make([]T, 0, c.chunkSize)
			}
		case item := <-c.incomingChan:
			chunk = append(chunk, item)
			if len(chunk) >= c.chunkSize {
				c.outgoingChan <- chunk
				chunk = make([]T, 0, c.chunkSize)
			}
		}
	}
}

type ChunkChanBuilder[T any] struct {
	ctx                context.Context
	chunkSize          int
	chunkTimeout       time.Duration
	incomingBufferSize int
	outgoingBufferSize int
	incomingChan       chan T
	chunkFunc          ChunkFunc[T]
}

func NewChunkChan[T any]() *ChunkChanBuilder[T] {
	return &ChunkChanBuilder[T]{
		ctx:                context.Background(),
		chunkSize:          100,
		chunkTimeout:       50 * time.Millisecond,
		incomingBufferSize: 1000,
		outgoingBufferSize: 100,
	}
}

func (b *ChunkChanBuilder[T]) WithContext(ctx context.Context) *ChunkChanBuilder[T] {
	b.ctx = ctx
	return b
}

func (b *ChunkChanBuilder[T]) WithChunkSize(chunkSize int) *ChunkChanBuilder[T] {
	b.chunkSize = chunkSize
	return b
}

func (b *ChunkChanBuilder[T]) WithChunkFunc(chunkFunc ChunkFunc[T]) *ChunkChanBuilder[T] {
	b.chunkFunc = chunkFunc
	return b
}

func (b *ChunkChanBuilder[T]) WithChunkTimeout(chunkTimeout time.Duration) *ChunkChanBuilder[T] {
	b.chunkTimeout = chunkTimeout
	return b
}

func (b *ChunkChanBuilder[T]) WithIncomingBufferSize(incomingBufferSize int) *ChunkChanBuilder[T] {
	b.incomingBufferSize = incomingBufferSize
	return b
}

func (b *ChunkChanBuilder[T]) WithOutgoingBufferSize(outgoingBufferSize int) *ChunkChanBuilder[T] {
	b.outgoingBufferSize = outgoingBufferSize
	return b
}

func (b *ChunkChanBuilder[T]) WithIncomingChan(incomingChan chan T) *ChunkChanBuilder[T] {
	b.incomingChan = incomingChan
	return b
}

func (b *ChunkChanBuilder[T]) Build() *ChunkChan[T] {
	if b.incomingChan == nil {
		b.incomingChan = make(chan T, b.incomingBufferSize)
	}
	res := &ChunkChan[T]{
		ctx:          b.ctx,
		chunkSize:    b.chunkSize,
		chunkTimeout: b.chunkTimeout,
		incomingChan: b.incomingChan,
		outgoingChan: make(chan []T, b.outgoingBufferSize),
	}

	go res.run()
	if b.chunkFunc != nil {
		go func() {
			for {
				select {
				case <-b.ctx.Done():
					return
				case chunk := <-res.outgoingChan:
					b.chunkFunc(chunk)
				}
			}
		}()
	}
	return res
}
