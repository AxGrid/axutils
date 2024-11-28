package collections

import (
	"context"
	"github.com/rs/zerolog"
	"sync"
	"time"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (11.04.2024)
*/

type ResponseMapBuilder[K comparable, V any] struct {
	ctx             context.Context
	logger          zerolog.Logger
	responseTimeout time.Duration
	clearTimeout    time.Duration
}

func NewResponseMap[K comparable, V any](ctx context.Context) *ResponseMapBuilder[K, V] {
	return &ResponseMapBuilder[K, V]{
		ctx:             ctx,
		responseTimeout: time.Second * 100,
		clearTimeout:    time.Second * 100,
	}
}

func (b *ResponseMapBuilder[K, V]) WithLogger(l zerolog.Logger) *ResponseMapBuilder[K, V] {
	b.logger = l
	return b
}

func (b *ResponseMapBuilder[K, V]) WithResponseTimeout(timeout time.Duration) *ResponseMapBuilder[K, V] {
	b.responseTimeout = timeout
	return b
}

func (b *ResponseMapBuilder[K, V]) WithClearTimeout(timeout time.Duration) *ResponseMapBuilder[K, V] {
	b.clearTimeout = timeout
	return b
}

func (b *ResponseMapBuilder[K, V]) Build() *ResponseMap[K, V] {
	rm := &ResponseMap[K, V]{
		logger:          b.logger,
		responseTimeout: b.responseTimeout,
		clearTimeout:    b.clearTimeout,
		mu:              sync.RWMutex{},
		m:               make(map[K]*chansHolder[K, V]),
	}
	go rm.clear(b.ctx)
	return rm
}

type ResponseMap[K comparable, V any] struct {
	logger          zerolog.Logger
	responseTimeout time.Duration
	clearTimeout    time.Duration
	mu              sync.RWMutex
	m               map[K]*chansHolder[K, V]
}

type chansHolder[K comparable, V any] struct {
	ctx       context.Context
	cancelCtx context.CancelFunc
	trx       K
	t         *time.Timer
	createdAt time.Time
	mu        sync.RWMutex
	isExist   bool
	data      V
	dataCh    chan V
	listeners []chan V
}

func newChansHolder[K comparable, V any](trx K, timeout time.Duration) *chansHolder[K, V] {
	ctx, cancel := context.WithCancel(context.Background())
	h := &chansHolder[K, V]{
		ctx:       ctx,
		cancelCtx: cancel,
		trx:       trx,
		t:         time.NewTimer(timeout),
		createdAt: time.Now(),
		dataCh:    make(chan V, 1),
	}
	go func() {
		defer h.t.Stop()
		for {
			select {
			case <-h.ctx.Done():
				close(h.dataCh)
				for _, ch := range h.listeners {
					close(ch)
				}
				return
			case <-h.t.C:
				for _, ch := range h.listeners {
					close(ch)
				}
				h.listeners = nil
			case d := <-h.dataCh:
				h.mu.Lock()
				h.data = d
				h.isExist = true
				for _, ch := range h.listeners {
					ch <- d
				}
				h.mu.Unlock()
			}
		}
	}()
	return h
}

func (h *chansHolder[K, V]) set(data V) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.isExist {
		return
	}
	h.dataCh <- data
}

func (h *chansHolder[K, V]) wait() V {
	h.mu.RLock()
	e := h.isExist
	h.mu.RUnlock()
	if e {
		return h.data
	}
	h.mu.Lock()
	e = h.isExist
	if e {
		h.mu.Unlock()
		return h.data
	}
	ch := make(chan V, 1)
	h.listeners = append(h.listeners, ch)
	h.mu.Unlock()
	return <-ch
}

func (r *ResponseMap[K, V]) Set(key K, value V) {
	holder := r.getHolder(key)
	holder.set(value)
}

func (r *ResponseMap[K, V]) Wait(key K) V {
	holder := r.getHolder(key)
	return holder.wait()
}

func (r *ResponseMap[K, V]) getHolder(key K) *chansHolder[K, V] {
	r.mu.RLock()
	holder, ok := r.m[key]
	r.mu.RUnlock()
	if ok {
		return holder
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	holder, ok = r.m[key]
	if ok {
		return holder
	}
	holder = newChansHolder[K, V](key, r.responseTimeout)
	r.m[key] = holder
	return holder
}

func (r *ResponseMap[K, V]) clear(ctx context.Context) {
	t := time.NewTicker(r.clearTimeout)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.logger.Debug().Int("holders count", len(r.m)).Msg("start cleaning response map")
			var remove []*chansHolder[K, V]
			r.mu.RLock()
			for _, holder := range r.m {
				if time.Now().After(holder.createdAt.Add(r.clearTimeout)) {
					remove = append(remove, holder)
				}
			}
			r.mu.RUnlock()
			r.logger.Debug().Int("should remove", len(remove)).Msg("collect expired holders")
			r.mu.Lock()
			for _, holder := range remove {
				holder.cancelCtx()
				delete(r.m, holder.trx)
			}
			r.mu.Unlock()
			r.logger.Debug().Int("holders count", len(r.m)).Msg("successfully cleaned expired holders")
		}
	}
}
