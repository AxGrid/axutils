package collections

import "sync"

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (11.04.2024)
*/

type ShardMap[K comparable, V any] struct {
	shards      []map[K]V
	shardsLock  []sync.RWMutex
	shardsFunc  func(K) int
	shardsCount int
}

func (m *ShardMap[K, V]) Get(key K) (V, bool) {
	shard := m.shardsFunc(key)
	m.shardsLock[shard].RLock()
	defer m.shardsLock[shard].RUnlock()
	v, ok := m.shards[shard][key]
	return v, ok
}

func (m *ShardMap[K, V]) Set(key K, value V) {
	shard := m.shardsFunc(key)
	m.shardsLock[shard].Lock()
	defer m.shardsLock[shard].Unlock()
	m.shards[shard][key] = value
}

func (m *ShardMap[K, V]) Delete(key K) {
	shard := m.shardsFunc(key)
	m.shardsLock[shard].Lock()
	defer m.shardsLock[shard].Unlock()
	delete(m.shards[shard], key)
}

func (m *ShardMap[K, V]) Has(key K) bool {
	shard := m.shardsFunc(key)
	m.shardsLock[shard].RLock()
	defer m.shardsLock[shard].RUnlock()
	_, ok := m.shards[shard][key]
	return ok
}

func (m *ShardMap[K, V]) SetIfNotExists(key K, value V) bool {
	shard := m.shardsFunc(key)
	m.shardsLock[shard].Lock()
	defer m.shardsLock[shard].Unlock()
	if _, ok := m.shards[shard][key]; ok {
		return false
	}
	m.shards[shard][key] = value
	return true
}

func (m *ShardMap[K, V]) Clear() {
	for i := 0; i < m.shardsCount; i++ {
		m.shardsLock[i].Lock()
		m.shards[i] = make(map[K]V)
		m.shardsLock[i].Unlock()
	}
}

func (m *ShardMap[K, V]) Size() int {
	count := 0
	for i := 0; i < m.shardsCount; i++ {
		m.shardsLock[i].RLock()
		count += len(m.shards[i])
		m.shardsLock[i].RUnlock()
	}
	return count
}

type ShardMapBuilder[K comparable, V any] struct {
	shardsCount int
	shardsFunc  func(K) int
}

func NewShardMap[K comparable, V any]() *ShardMapBuilder[K, V] {
	return &ShardMapBuilder[K, V]{
		shardsCount: 6,
	}
}

func (b *ShardMapBuilder[K, V]) WithShardsCount(shardsCount int) *ShardMapBuilder[K, V] {
	b.shardsCount = shardsCount
	return b
}

func (b *ShardMapBuilder[K, V]) WithShardsFunc(shardsFunc func(K) int) *ShardMapBuilder[K, V] {
	b.shardsFunc = shardsFunc
	return b
}

func (b *ShardMapBuilder[K, V]) Build() *ShardMap[K, V] {
	//если K - numeric то shardsFunc можно назначить автоматом
	m := &ShardMap[K, V]{
		shards:      make([]map[K]V, b.shardsCount),
		shardsLock:  make([]sync.RWMutex, b.shardsCount),
		shardsFunc:  b.shardsFunc,
		shardsCount: b.shardsCount,
	}
	for i := 0; i < b.shardsCount; i++ {
		m.shards[i] = make(map[K]V)
	}
	return m
}
