package collections

import (
	"github.com/stretchr/testify/assert"
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
	buildCount := 0
	unloadCount := 0
	m := NewGuavaMap[int, int]().WithMaxCount(10).WithLoadFunc(func(key int) (int, error) {
		buildCount++
		return key * 10, nil
	}).WithUnloadFunc(func(key int, value int) {
		unloadCount++
	}).
		WithMaxCount(50).Build()
	for i := 0; i < 100; i++ {
		v, err := m.Get(i)
		assert.Nil(t, err)
		assert.Equal(t, i*10, v)
	}
	assert.Equal(t, 50, m.Size())
	assert.Equal(t, 100, buildCount)
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, 50, unloadCount)
}

func TestGuavaMap_GetWithWriteTimeout(t *testing.T) {
	buildCount := 0
	unloadCount := 0
	m := NewGuavaMap[int, int]().WithMaxCount(10).WithLoadFunc(func(key int) (int, error) {
		buildCount++
		return key * 10, nil
	}).WithUnloadFunc(func(key int, value int) {
		unloadCount++
	}).
		WithMaxCount(50).WithWriteTimeout(time.Millisecond * 20).WithClearTimeout(time.Millisecond * 10).Build()
	for i := 0; i < 100; i++ {
		v, err := m.Get(i)
		assert.Nil(t, err)
		assert.Equal(t, i*10, v)
	}
	assert.Equal(t, 50, m.Size())
	assert.Equal(t, 100, buildCount)
	time.Sleep(time.Millisecond * 1000)
	assert.Equal(t, 100, unloadCount)
	assert.Equal(t, 0, m.Size())

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
