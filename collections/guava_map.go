package collections

import (
	"context"
	"sync"
	"time"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (14.03.2024)
*/

type GuavaLoadFunc[K comparable, V any] func(K) (V, error)
type GuavaUnloadFunc[K comparable, V any] func(K, V)

type timeoutHolder[K comparable] struct {
	key  K
	time time.Time
}

type GuavaMap[K comparable, V any] struct {
	stored             map[K]V
	storedSlice        []K
	mu                 sync.RWMutex
	maxCount           int
	loadFunc           GuavaLoadFunc[K, V]
	unloadFunc         GuavaUnloadFunc[K, V]
	enableWriteTimeout bool
	writeTimeout       time.Duration
	timeoutSlices      []*timeoutHolder[K]
	timeoutMu          sync.Mutex
	clearTimeout       time.Duration
	lockLoad           *MapMutex[K]
	ctx                context.Context
}

func (m *GuavaMap[K, V]) timeoutThreadFunction() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-time.After(m.clearTimeout):
			m.timeoutMu.Lock()
			deleteList := make([]K, 0)
			for len(m.timeoutSlices) > 0 && time.Since(m.timeoutSlices[0].time) >= m.writeTimeout {
				key := m.timeoutSlices[0].key
				m.timeoutSlices = m.timeoutSlices[1:]
				deleteList = append(deleteList, key)
			}
			m.timeoutMu.Unlock()
			for _, key := range deleteList {
				m.Delete(key)
			}
		}
	}
}

func (m *GuavaMap[K, V]) Has(key K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.stored[key]
	return ok
}

func (m *GuavaMap[K, V]) HasOrCreate(key K, value V) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.stored[key]
	if !ok {
		m.stored[key] = value
	}
	return ok
}

func (m *GuavaMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.unloadFunc != nil {
		go m.unloadFunc(key, m.stored[key])
	}
	delete(m.stored, key)
	if m.maxCount > 0 {
		for i, k := range m.storedSlice {
			if k == key {
				m.storedSlice = append(m.storedSlice[:i], m.storedSlice[i+1:]...)
				break
			}
		}
	}
	if m.enableWriteTimeout {
		m.timeoutMu.Lock()
		for i, k := range m.timeoutSlices {
			if k.key == key {
				m.timeoutSlices = append(m.timeoutSlices[:i], m.timeoutSlices[i+1:]...)
				break
			}
		}
		m.timeoutMu.Unlock()
	}
}

func (m *GuavaMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.unloadFunc != nil {
		for k, v := range m.stored {
			go m.unloadFunc(k, v)
		}
	}
	m.stored = make(map[K]V)
	if m.maxCount > 0 {
		m.storedSlice = make([]K, 0, m.maxCount)
	}
	if m.enableWriteTimeout {
		m.timeoutMu.Lock()
		defer m.timeoutMu.Unlock()
		m.timeoutSlices = make([]*timeoutHolder[K], 0)
	}

}

func (m *GuavaMap[K, V]) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.stored)
}

func (m *GuavaMap[K, V]) Get(key K) (V, error) {
	m.mu.RLock()
	val, ok := m.stored[key]
	m.mu.RUnlock()
	if ok {
		return val, nil
	}
	if m.loadFunc == nil {
		return val, nil
	}
	var err error
	var loadedVal V

	if m.lockLoad != nil {
		m.lockLoad.Lock(key)
		if !m.Has(key) {
			loadedVal, err = m.loadFunc(key)
		}
		m.lockLoad.Unlock(key)
	} else {
		loadedVal, err = m.loadFunc(key)
	}

	if err != nil {
		return val, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok = m.stored[key]
	if ok {
		return val, nil
	}
	if m.maxCount > 0 {
		if len(m.storedSlice) >= m.maxCount {
			if m.unloadFunc != nil {
				go m.unloadFunc(m.storedSlice[0], m.stored[m.storedSlice[0]])
			}
			delete(m.stored, m.storedSlice[0])
			m.storedSlice = m.storedSlice[1:]
			if m.enableWriteTimeout {
				m.timeoutMu.Lock()
				m.timeoutSlices = m.timeoutSlices[1:]
				m.timeoutMu.Unlock()
			}
		}
		m.storedSlice = append(m.storedSlice, key)
	}
	if m.enableWriteTimeout {
		m.timeoutMu.Lock()
		m.timeoutSlices = append(m.timeoutSlices, &timeoutHolder[K]{key: key, time: time.Now()})
		m.timeoutMu.Unlock()
	}
	m.stored[key] = loadedVal
	return loadedVal, nil
}

func (m *GuavaMap[K, V]) Set(key K, val V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.stored[key]
	m.stored[key] = val
	if ok {
		return
	}
	if m.maxCount > 0 {
		if len(m.storedSlice) >= m.maxCount {
			if m.unloadFunc != nil {
				go m.unloadFunc(m.storedSlice[0], m.stored[m.storedSlice[0]])
			}
			delete(m.stored, m.storedSlice[0])
			m.storedSlice = m.storedSlice[1:]
			if m.enableWriteTimeout {
				m.timeoutMu.Lock()
				m.timeoutSlices = m.timeoutSlices[1:]
				m.timeoutMu.Unlock()
			}
		}
		m.storedSlice = append(m.storedSlice, key)
	}
	if m.enableWriteTimeout {
		m.timeoutMu.Lock()
		m.timeoutSlices = append(m.timeoutSlices, &timeoutHolder[K]{key: key, time: time.Now()})
		m.timeoutMu.Unlock()
	}
}

type GuavaMapBuilder[K comparable, V any] struct {
	loadFunc           GuavaLoadFunc[K, V]
	lockLoad           bool
	maxCount           int
	unloadFunc         GuavaUnloadFunc[K, V]
	writeTimeout       time.Duration
	enableWriteTimeout bool
	ctx                context.Context
	clearTimeout       time.Duration
}

func NewGuavaMap[K comparable, V any]() *GuavaMapBuilder[K, V] {
	return &GuavaMapBuilder[K, V]{
		ctx:          context.Background(),
		clearTimeout: time.Second * 5,
	}
}

func (b *GuavaMapBuilder[K, V]) WithContext(ctx context.Context) *GuavaMapBuilder[K, V] {
	b.ctx = ctx
	return b
}

func (b *GuavaMapBuilder[K, V]) WithLockLoad(lockLoad bool) *GuavaMapBuilder[K, V] {
	b.lockLoad = lockLoad
	return b
}

func (b *GuavaMapBuilder[K, V]) WithClearTimeout(clearTimeout time.Duration) *GuavaMapBuilder[K, V] {
	b.clearTimeout = clearTimeout
	return b
}

func (b *GuavaMapBuilder[K, V]) WithLoadFunc(loadFunc GuavaLoadFunc[K, V]) *GuavaMapBuilder[K, V] {
	b.loadFunc = loadFunc
	return b
}

func (b *GuavaMapBuilder[K, V]) WithMaxCount(maxCount int) *GuavaMapBuilder[K, V] {
	b.maxCount = maxCount
	return b
}

func (b *GuavaMapBuilder[K, V]) WithUnloadFunc(unloadFunc GuavaUnloadFunc[K, V]) *GuavaMapBuilder[K, V] {
	b.unloadFunc = unloadFunc
	return b
}

func (b *GuavaMapBuilder[K, V]) WithWriteTimeout(writeTimeout time.Duration) *GuavaMapBuilder[K, V] {
	b.writeTimeout = writeTimeout
	b.enableWriteTimeout = b.writeTimeout > 0
	return b
}

func (b *GuavaMapBuilder[K, V]) Build() *GuavaMap[K, V] {
	res := &GuavaMap[K, V]{
		stored:             make(map[K]V),
		loadFunc:           b.loadFunc,
		maxCount:           b.maxCount,
		unloadFunc:         b.unloadFunc,
		writeTimeout:       b.writeTimeout,
		enableWriteTimeout: b.enableWriteTimeout,
		ctx:                b.ctx,
	}
	if b.lockLoad {
		res.lockLoad = NewMapMutex[K]()
	}
	if b.maxCount > 0 {
		res.storedSlice = make([]K, 0, b.maxCount)
	}
	if b.enableWriteTimeout {
		go res.timeoutThreadFunction()
	}
	return res
}
