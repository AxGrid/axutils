package collections

import (
	"fmt"
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
zed (14.03.2024)
*/

func TestGuavaMap_Get(t *testing.T) {
	buildCount := 0
	m := NewGuavaMap[int, int]().WithMaxCount(10).WithLoadFunc(func(key int) (int, error) {
		buildCount++
		return key * 10, nil
	}).Build()
	v, err := m.Get(10)
	assert.Nil(t, err)
	assert.Equal(t, 100, v)
	assert.Equal(t, 1, buildCount)
	v, err = m.Get(10)
	assert.Nil(t, err)
	assert.Equal(t, 100, v)
	assert.Equal(t, 1, buildCount)
	assert.Equal(t, 1, m.Size())
	m.Set(15, 500)
	assert.Equal(t, 2, m.Size())
	assert.Equal(t, 1, buildCount)
	v, err = m.Get(15)
	assert.Nil(t, err)
	assert.Equal(t, 500, v)
}

func TestGuavaMap_HiLoad(t *testing.T) {
	buildCount := 0
	m := NewGuavaMap[int, int]().WithMaxCount(10).WithLoadFunc(func(key int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		buildCount++
		return key * 10, nil
	}).Build()
	go func() {
		v, err := m.Get(10)
		assert.Nil(t, err)
		assert.Equal(t, 100, v)
	}()

	go func() {
		v, err := m.Get(15)
		assert.Nil(t, err)
		assert.Equal(t, 150, v)
	}()

	go func() {
		v, err := m.Get(10)
		assert.Nil(t, err)
		assert.Equal(t, 100, v)
	}()
	time.Sleep(250 * time.Millisecond)
	assert.Equal(t, 3, buildCount)
	v, err := m.Get(10)
	assert.Nil(t, err)
	assert.Equal(t, 100, v)
	assert.Equal(t, 3, buildCount)
}

func TestGuavaMap_GetWithMaxCount(t *testing.T) {
	buildCount := int32(0)
	unloadCount := int32(0)
	m := NewGuavaMap[int, int]().WithLoadFunc(func(key int) (int, error) {
		atomic.AddInt32(&buildCount, 1)
		return key * 10, nil
	}).WithUnloadFunc(func(key int, value int) {
		atomic.AddInt32(&unloadCount, 1)
	}).
		WithMaxCount(50).Build()

	for i := 0; i < 100; i++ {
		v, err := m.Get(i)
		assert.Nil(t, err)
		assert.Equal(t, i*10, v)
	}

	assert.Equal(t, 50, m.Size())
	assert.Equal(t, 100, int(buildCount))
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, 50, int(unloadCount))
}

func TestGuavaMap_SetWithMaxSize(t *testing.T) {

	buildCount := int32(0)
	unloadCount := int32(0)
	m := NewGuavaMap[int, int]().WithMaxCount(50).WithWriteTimeout(time.Millisecond * 200).
		WithLoadFunc(func(key int) (int, error) {
			atomic.AddInt32(&buildCount, 1)
			return key * 10, nil
		}).WithUnloadFunc(func(key int, value int) {
		atomic.AddInt32(&unloadCount, 1)
	}).Build()
	for i := 0; i < 100; i++ {
		m.Set(i, i*10)
	}
	assert.Equal(t, 50, m.Size())
	assert.Equal(t, 0, int(buildCount))
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, 50, int(unloadCount))
	assert.Equal(t, 50, m.Size())
}

func TestGuavaMap_GetAndDelete(t *testing.T) {
	unload := int32(0)
	load := int32(0)
	g := NewGuavaMap[int, int]().WithMaxCount(100).WithLoadFunc(func(key int) (int, error) {
		atomic.AddInt32(&load, 1)
		return key * 10, nil
	}).WithUnloadFunc(func(k int, v int) {
		atomic.AddInt32(&unload, 1)
	}).Build()
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(_i int) {
			v, err := g.Get(_i)
			assert.Nil(t, err)
			assert.Equal(t, _i*10, v)
			wg.Done()
		}(i)
	}
	wg.Wait()
	time.Sleep(time.Millisecond * 10)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(_i int) {
			g.Delete(_i)
			wg.Done()
		}(i)
	}
	wg.Wait()
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, 0, g.Size())
	assert.Equal(t, 100, int(load))
	assert.Equal(t, 100, int(unload))
}

func TestGuavaMap_GetWithWriteTimeout(t *testing.T) {
	buildCount := int32(0)
	unloadCount := int32(0)
	m := NewGuavaMap[int, int]().WithMaxCount(100).WithLoadFunc(func(key int) (int, error) {
		atomic.AddInt32(&buildCount, 1)
		return key * 10, nil
	}).WithUnloadFunc(func(key int, value int) {
		atomic.AddInt32(&unloadCount, 1)
	}).
		WithWriteTimeout(time.Millisecond * 20).
		Build()
	for i := 0; i < 100; i++ {
		v, err := m.Get(i)
		assert.Nil(t, err)
		assert.Equal(t, i*10, v)
	}
	assert.Equal(t, 100, m.Size())
	assert.Equal(t, 100, int(buildCount))
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, 0, m.Size())
	assert.Equal(t, 100, int(unloadCount))
}

func TestGuavaMapBuilder_WithReadTimeout(t *testing.T) {
	buildCount := int32(0)
	unloadCount := int32(0)
	m := NewGuavaMap[int, int]().WithMaxCount(100).WithLoadFunc(func(key int) (int, error) {
		atomic.AddInt32(&buildCount, 1)
		return key * 10, nil
	}).WithUnloadFunc(func(key int, value int) {
		atomic.AddInt32(&unloadCount, 1)
	}).
		WithReadTimeout(time.Millisecond * 20).
		Build()
	for i := 0; i < 100; i++ {
		v, err := m.Get(i)
		assert.Nil(t, err)
		assert.Equal(t, i*10, v)
	}
	assert.Equal(t, 100, m.Size())
	assert.Equal(t, 100, int(buildCount))

	for j := 0; j < 5; j++ {
		time.Sleep(time.Millisecond * 10)
		for i := 0; i < 50; i++ {
			v, err := m.Get(i)
			assert.Nil(t, err)
			assert.Equal(t, i*10, v)
		}
	}
	assert.Equal(t, 100, int(buildCount))
	assert.Equal(t, 50, int(unloadCount))
	assert.Equal(t, 50, m.Size())
	time.Sleep(time.Millisecond * 300)
	assert.Equal(t, 0, m.Size())
	assert.Equal(t, 100, int(unloadCount))

}

func TestGuavaMap_LockLoad(t *testing.T) {
	buildCount := 0
	m := NewGuavaMap[int, int]().WithMaxCount(10).WithLockLoad(true).WithLoadFunc(func(key int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		buildCount++
		return key * 10, nil
	}).Build()
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			_, err := m.Get(10)
			assert.Nil(t, err)
			wg.Done()
		}()
	}
	wg.Wait()
	assert.Equal(t, 1, buildCount)
}

func BenchmarkGuavaMap_Get(b *testing.B) {
	m := NewGuavaMap[int, int]().WithMaxCount(10).WithLoadFunc(func(key int) (int, error) {
		return key * 10, nil
	}).Build()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(i)
	}
}

func BenchmarkGuavaMap_GetWithMaxCount(b *testing.B) {
	m := NewGuavaMap[int, int]().WithMaxCount(10).WithLoadFunc(func(key int) (int, error) {
		return key * 10, nil
	}).WithMaxCount(50).Build()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(i)
	}
}

func BenchmarkGuavaMap_LockLoad(b *testing.B) {
	loadCount := int32(0)
	m := NewGuavaMap[int, int]().WithLoadFunc(func(key int) (int, error) {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&loadCount, 1)
		return key*10 + 1, nil
	}).WithLockLoad(true).Build()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val, err := m.Get(i % 20)
		if err != nil {
			b.Fatal()
		}
		if val == 0 {
			b.Fatal()
		}
	}
	println("loadCount", loadCount)
}

func TestGuavaMap_HasOrCreate(t *testing.T) {
	m := NewGuavaMap[string, bool]().WithMaxCount(1000).Build()
	for i := 0; i < 10000; i++ {
		v := m.HasOrCreate(fmt.Sprintf("key-%d", i), true)
		assert.False(t, v)
	}
	assert.Equal(t, 1000, m.Size())
}
