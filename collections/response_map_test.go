package collections

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"os"
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
	m := NewResponseMap[string, *mockResponse](context.Background()).Build()
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
	m := NewResponseMap[string, *mockResponse](context.Background()).WithResponseTimeout(timeout).Build()
	k := "key"
	v := m.Wait(k)
	assert.Nil(t, v)
}

func TestResponseMap_WaitMultiple(t *testing.T) {
	m := NewResponseMap[string, *mockResponse](context.Background()).Build()
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
	m := NewResponseMap[string, *mockResponse](context.Background()).WithResponseTimeout(timeout).Build()
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
	m := NewResponseMap[string, *mockResponse](context.Background()).Build()
	k := "key"
	res := []byte("ok")

	m.Set(k, &mockResponse{data: res})

	v := m.Wait(k)
	assert.NotNil(t, v)
	assert.Nil(t, v.err)
	assert.Equal(t, res, v.data)
}

func TestResponseMap_ClearTimeout_Wait(t *testing.T) {
	l := zerolog.New(os.Stdout)
	m := NewResponseMap[string, *mockResponse](context.Background()).WithLogger(l).WithClearTimeout(1 * time.Second).Build()
	k := "key"
	v := m.Wait(k)
	assert.Nil(t, v)
	time.Sleep(1500 * time.Millisecond)
	assert.Empty(t, m.m)
}

func TestResponseMap_ClearTimeout_Set(t *testing.T) {
	l := zerolog.New(os.Stdout)
	m := NewResponseMap[string, *mockResponse](context.Background()).WithClearTimeout(2 * time.Second).WithLogger(l).Build()
	k := "key"
	go func() {
		for i := 0; i < 100; i++ {
			k = fmt.Sprintf("key-%d", i)
			m.Set(k, &mockResponse{data: []byte("ok")})
		}
	}()
	time.Sleep(4500 * time.Millisecond)
	assert.Empty(t, m.m)
}
