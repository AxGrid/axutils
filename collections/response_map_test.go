package collections

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type mockResponse struct {
	data []byte
	err  error
}

func TestResponseMap_Wait(t *testing.T) {
	m := NewResponseMap[string, *mockResponse]().Build()
	k := "key"
	res := []byte("ok")
	go func() {
		v := m.Wait(k)
		assert.Nil(t, v.err)
		assert.Equal(t, v.data, res)
	}()
	time.Sleep(1 * time.Second)
	m.Set(k, &mockResponse{data: res})
	time.Sleep(1 * time.Second)
}

func TestResponseMap_WaitTimeout(t *testing.T) {
	timeout := 2 * time.Second
	m := NewResponseMap[string, *mockResponse]().WithTimeout(timeout).Build()
	k := "key"
	v := m.Wait(k)
	assert.Nil(t, v)
}

func TestResponseMap_WaitMultiple(t *testing.T) {
	m := NewResponseMap[string, *mockResponse]().Build()
	k := "key"
	res := []byte("ok")

	wg := sync.WaitGroup{}
	var shouldResps, resps uint64 = 100000, 0
	for i := 0; i < int(shouldResps); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := m.Wait(k)
			assert.Nil(t, v.err)
			assert.Equal(t, v.data, res)
			atomic.AddUint64(&resps, 1)
		}()
	}
	time.Sleep(1 * time.Second)
	m.Set(k, &mockResponse{data: res})
	wg.Wait()
	assert.Equal(t, shouldResps, resps)
}

func TestResponseMap_WaitMultipleTimeout(t *testing.T) {
	timeout := 2 * time.Second
	m := NewResponseMap[string, *mockResponse]().WithTimeout(timeout).Build()
	k := "key"

	wg := sync.WaitGroup{}
	var shouldResps, resps uint64 = 100000, 0
	for i := 0; i < int(shouldResps); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := m.Wait(k)
			assert.Nil(t, v)
			atomic.AddUint64(&resps, 1)
		}()
	}
	wg.Wait()
	assert.Equal(t, shouldResps, resps)
}

func TestResponseMap_Set(t *testing.T) {
	m := NewResponseMap[string, *mockResponse]().Build()
	k := "key"
	res := []byte("ok")

	m.Set(k, &mockResponse{data: res})

	v := m.Wait(k)
	assert.NotNil(t, v)
	assert.Nil(t, v.err)
	assert.Equal(t, res, v.data)
}
