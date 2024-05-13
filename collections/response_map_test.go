package collections

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type mockResponse struct {
	data []byte
	err  error
}

//func TestResponseMap_Wait(t *testing.T) {
//	m := NewResponseMap[uint64, *mockResponse](1)
//
//	go func() {
//		time.Sleep(2 * time.Second)
//		m.Set(1, &mockResponse{data: []byte("1")})
//	}()
//
//	res := m.Wait(1)
//	assert.Equal(t, []byte("1"), res.data)
//	assert.Nil(t, res.err)
//}
//
//func TestResponseMap_Wait_Timeout(t *testing.T) {
//	m := NewResponseMap[uint64, *mockResponse](1)
//
//	go func() {
//		time.Sleep(3 * time.Second)
//		m.Set(1, &mockResponse{data: []byte("1")})
//	}()
//
//	res := m.Wait(1)
//	assert.Nil(t, res)
//}

func TestResponseMap_Wait(t *testing.T) {
	m := NewResponseMap[uint64, *mockResponse]().Build()

	go func() {
		time.Sleep(2 * time.Second)
		m.Set(1, &mockResponse{data: []byte("1")})
	}()

	res := m.Wait(1)
	assert.Equal(t, []byte("1"), res.data)
	assert.Nil(t, res.err)
}

func TestResponseMap_Wait_Timeout(t *testing.T) {
	m := NewResponseMap[uint64, *mockResponse]().WithTimeout(1).Build()

	go func() {
		time.Sleep(3 * time.Second)
		m.Set(1, &mockResponse{data: []byte("1")})
	}()

	res := m.Wait(1)
	assert.Nil(t, res)
}

func TestResponseMap_Wait_Multiple_Keys(t *testing.T) {
	m := NewResponseMap[uint64, *mockResponse]().Build()

	go func() {
		time.Sleep(2 * time.Second)
		assert.Equal(t, 100, len(m.m[1].chans))
	}()

	go func() {
		time.Sleep(3 * time.Second)
		m.Set(1, &mockResponse{data: []byte("1")})
	}()

	for i := 0; i < 100; i++ {
		go func() {
			res := m.Wait(1)
			assert.NotNil(t, res)
			assert.Equal(t, []byte("1"), res.data)
			assert.Nil(t, res.err)
		}()
	}

	time.Sleep(5 * time.Second)
}

func TestResponseMap_Wait_Multiple_Keys_Timeout(t *testing.T) {
	m := NewResponseMap[uint64, *mockResponse]().WithTimeout(1).Build()

	go func() {
		time.Sleep(3 * time.Second)
		m.Set(1, &mockResponse{data: []byte("1")})
	}()

	shouldResps := 100
	resps := 0
	for i := 0; i < shouldResps; i++ {
		go func() {
			res := m.Wait(1)
			assert.Nil(t, res)
			resps++
		}()
	}
	time.Sleep(4 * time.Second)
	assert.Equal(t, shouldResps, resps)
}
