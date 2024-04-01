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

type guavaHolder[V any] struct {
	timer    *time.Timer
	cancelFn context.CancelFunc
	v        V
}

type GuavaMap[K comparable, V any] struct {
	stored             map[K]*guavaHolder[V]
	storedSlice        []K
	updateLockMap      map[K]*sync.Mutex
	updateLockMapMu    sync.RWMutex
	mu                 sync.RWMutex
	maxCount           int
	loadFunc           GuavaLoadFunc[K, V]
	unloadFunc         GuavaUnloadFunc[K, V]
	enableWriteTimeout bool
	writeTimeout       time.Duration
	readTimeout        time.Duration
	lockLoad           *MapMutex[K]
	ctx                context.Context
}

func (m *GuavaMap[K, V]) getKeyLock(key K) *sync.Mutex {
	m.updateLockMapMu.Lock()
	defer m.updateLockMapMu.Unlock()
	lock, ok := m.updateLockMap[key]
	if !ok {
		lock = &sync.Mutex{}
		m.updateLockMap[key] = lock
	}
	return lock
}

func (m *GuavaMap[K, V]) createHolder(key K, value V) *guavaHolder[V] {
	res := &guavaHolder[V]{
		v: value,
	}
	if m.readTimeout > 0 || m.writeTimeout > 0 {
		ctx, cancelFn := context.WithCancel(m.ctx)
		res.cancelFn = cancelFn
		if m.writeTimeout > 0 {
			res.timer = time.NewTimer(m.writeTimeout)
		} else {
			res.timer = time.NewTimer(m.readTimeout)
		}
		go func() {
			defer res.timer.Stop()
			select {
			case <-ctx.Done():
				return
			case <-res.timer.C:
				m.mu.Lock()
				m.safeDelete(key)
				m.mu.Unlock()

			}
		}()

	}
	return res
}

func (m *GuavaMap[K, V]) Has(key K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.stored[key]
	if ok {
		if v.timer != nil && m.readTimeout > 0 {
			v.timer.Reset(m.readTimeout)
		}
	}
	return ok
}

func (m *GuavaMap[K, V]) HasOrCreate(key K, value V) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.stored[key]
	if !ok {
		m.stored[key] = m.createHolder(key, value)
	} else {
		if m.stored[key].timer != nil && m.readTimeout > 0 {
			m.stored[key].timer.Reset(m.readTimeout)
		}
	}
	return ok
}

func (m *GuavaMap[K, V]) safeDelete(key K) {
	if m.unloadFunc != nil {
		v := m.stored[key].v
		go m.unloadFunc(key, v)
	}
	if m.readTimeout > 0 || m.writeTimeout > 0 {
		v := m.stored[key]
		if v.cancelFn != nil {
			v.cancelFn()
		}
	}

	delete(m.stored, key)
	m.updateLockMapMu.Lock()
	delete(m.updateLockMap, key)
	m.updateLockMapMu.Unlock()
	if m.maxCount > 0 {
		for i, k := range m.storedSlice {
			if k == key {
				m.storedSlice = append(m.storedSlice[:i], m.storedSlice[i+1:]...)
				break
			}
		}
	}

	if m.updateLockMap != nil {
		m.updateLockMapMu.RLock()
		_, ok := m.updateLockMap[key]
		m.updateLockMapMu.RUnlock()
		if ok {
			m.updateLockMapMu.Lock()
			delete(m.updateLockMap, key)
			m.updateLockMapMu.Unlock()
		}
	}
}

func (m *GuavaMap[K, V]) LockForUpdate(key K, update func() V) V {
	if m.updateLockMap == nil {
		m.updateLockMapMu.Lock()
		if m.updateLockMap == nil {
			m.updateLockMap = make(map[K]*sync.Mutex)
		}
		m.updateLockMapMu.Unlock()
	}
	m.updateLockMapMu.RLock()
	lock, ok := m.updateLockMap[key]
	m.updateLockMapMu.RUnlock()
	if !ok {
		m.updateLockMapMu.Lock()
		lock, ok = m.updateLockMap[key]
		if !ok {
			lock = &sync.Mutex{}
			m.updateLockMap[key] = lock
		}
		m.updateLockMapMu.Unlock()
	}
	lock.Lock()
	defer lock.Unlock()
	res := update()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stored[key] = m.createHolder(key, res)
	return res
}

func (m *GuavaMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.safeDelete(key)
}

func (m *GuavaMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.unloadFunc != nil {
		for k, v := range m.stored {
			go m.unloadFunc(k, v.v)
			if v.cancelFn != nil {
				v.cancelFn()
			}
		}
	}
	m.stored = make(map[K]*guavaHolder[V])
	if m.maxCount > 0 {
		m.storedSlice = make([]K, 0, m.maxCount)
	}
	if m.updateLockMap != nil {
		m.updateLockMapMu.Lock()
		defer m.updateLockMapMu.Unlock()
		m.updateLockMap = make(map[K]*sync.Mutex)
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
		if val.timer != nil && m.readTimeout > 0 {
			val.timer.Reset(m.readTimeout)
		}
		return val.v, nil
	}
	if m.loadFunc == nil {
		var v V
		return v, nil
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
		var v V
		return v, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok = m.stored[key]
	if ok {
		if val.timer != nil && m.readTimeout > 0 {
			val.timer.Reset(m.readTimeout)
		}
		return val.v, nil
	}
	if m.maxCount > 0 {
		if len(m.storedSlice) >= m.maxCount {
			m.safeDelete(m.storedSlice[0])
		}
		m.storedSlice = append(m.storedSlice, key)
	}
	m.stored[key] = m.createHolder(key, loadedVal)
	return loadedVal, nil
}

func (m *GuavaMap[K, V]) Set(key K, val V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.stored[key]
	if !ok {
		m.stored[key] = m.createHolder(key, val)
		if m.maxCount > 0 {
			if len(m.storedSlice) >= m.maxCount {
				m.safeDelete(m.storedSlice[0])
			}
			m.storedSlice = append(m.storedSlice, key)
		}
		return
	} else {
		if v.timer != nil {
			if m.writeTimeout > 0 {
				v.timer.Reset(m.writeTimeout)
			} else {
				v.timer.Reset(m.readTimeout)
			}
		}
		v.v = val
		return
	}

}

type GuavaMapBuilder[K comparable, V any] struct {
	loadFunc     GuavaLoadFunc[K, V]
	lockLoad     bool
	maxCount     int
	unloadFunc   GuavaUnloadFunc[K, V]
	writeTimeout time.Duration
	readTimeout  time.Duration
	ctx          context.Context
}

func NewGuavaMap[K comparable, V any]() *GuavaMapBuilder[K, V] {
	return &GuavaMapBuilder[K, V]{
		ctx: context.Background(),
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
	return b
}

func (b *GuavaMapBuilder[K, V]) WithReadTimeout(readTimeout time.Duration) *GuavaMapBuilder[K, V] {
	b.readTimeout = readTimeout
	return b
}

func (b *GuavaMapBuilder[K, V]) Build() *GuavaMap[K, V] {
	res := &GuavaMap[K, V]{
		stored:       make(map[K]*guavaHolder[V]),
		loadFunc:     b.loadFunc,
		maxCount:     b.maxCount,
		unloadFunc:   b.unloadFunc,
		writeTimeout: b.writeTimeout,
		readTimeout:  b.readTimeout,
		ctx:          b.ctx,
	}
	if b.lockLoad {
		res.lockLoad = NewMapMutex[K]()
	}
	if b.maxCount > 0 {
		res.storedSlice = make([]K, 0, b.maxCount)
	}
	return res
}
