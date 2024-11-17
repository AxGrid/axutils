package collections

import (
	"context"
	"sync"
	"time"
)

/*
   ________  ________   _______   ______    ______   _______    _______
  /        \/        \//       \//      \ //      \ /       \\//       \
 /        _/        _//        //       ///       //        ///        /
/-        //       //        _/        //        //         /        _/
\_______// \_____// \________/\________/\________/\___/____/\____/___/
zed (03.11.2024)
*/

type WaitMap[K comparable, V any] struct {
	waiterChannels map[K][]chan V
	dataHolder     map[K]V
	mu             sync.RWMutex
	requestTimeout time.Duration
	responseTtl    time.Duration
	ctx            context.Context
}

func (wm *WaitMap[K, V]) Count() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	a := len(wm.waiterChannels)
	b := len(wm.dataHolder)
	return a + b
}

func (wm *WaitMap[K, V]) Wait(key K) V {
	wm.mu.RLock()
	value, ok := wm.dataHolder[key]
	wm.mu.RUnlock()
	if ok {
		return value
	}

	wm.mu.Lock()
	value, ok = wm.dataHolder[key]
	if ok {
		wm.mu.Unlock()
		return value
	}

	ch := make(chan V, 1)
	_, ok = wm.waiterChannels[key]
	if !ok {
		go func() {
			select {
			case <-wm.ctx.Done():
				wm.mu.Lock()
				if chans, exists := wm.waiterChannels[key]; exists {
					var v V
					for _, ch := range chans {
						ch <- v
						close(ch)
					}
					delete(wm.waiterChannels, key)
				}
				wm.mu.Unlock()
			case <-time.After(wm.requestTimeout): // TIMEOUT
				var v V
				wm.Set(key, v)
			}
		}()
	}
	wm.waiterChannels[key] = append(wm.waiterChannels[key], ch)
	wm.mu.Unlock()

	value = <-ch
	return value
}

func (wm *WaitMap[K, V]) Set(key K, value V) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if _, ok := wm.dataHolder[key]; ok {
		return
	}

	wm.dataHolder[key] = value
	if chans, ok := wm.waiterChannels[key]; ok {
		for _, ch := range chans {
			ch <- value
			close(ch) // Закрываем канал после отправки значения
		}
		delete(wm.waiterChannels, key)
	}

	go func() {
		select {
		case <-wm.ctx.Done():
			wm.mu.Lock()
			delete(wm.dataHolder, key)
			wm.mu.Unlock()
			return
		case <-time.After(wm.responseTtl):
			wm.mu.Lock()
			delete(wm.dataHolder, key)
			wm.mu.Unlock()
		}
	}()
}

type WaitMapBuilder[K comparable, V any] struct {
	requestTimeout time.Duration
	responseTtl    time.Duration
	ctx            context.Context
}

func NewWaitMap[K comparable, V any]() *WaitMapBuilder[K, V] {
	return &WaitMapBuilder[K, V]{
		requestTimeout: time.Second * 10,
		responseTtl:    time.Minute * 5,
		ctx:            context.Background(),
	}
}

func (b *WaitMapBuilder[K, V]) WithRequestTimeout(timeout time.Duration) *WaitMapBuilder[K, V] {
	b.requestTimeout = timeout
	return b
}

func (b *WaitMapBuilder[K, V]) WithResponseTtl(ttl time.Duration) *WaitMapBuilder[K, V] {
	b.responseTtl = ttl
	return b
}

func (b *WaitMapBuilder[K, V]) WithContext(ctx context.Context) *WaitMapBuilder[K, V] {
	b.ctx = ctx
	return b
}

func (b *WaitMapBuilder[K, V]) Build() *WaitMap[K, V] {
	return &WaitMap[K, V]{
		waiterChannels: make(map[K][]chan V),
		dataHolder:     make(map[K]V),
		mu:             sync.RWMutex{},
		requestTimeout: b.requestTimeout,
		responseTtl:    b.responseTtl,
		ctx:            b.ctx,
	}
}
