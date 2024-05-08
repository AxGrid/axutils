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

func TestResponseMap_Wait(t *testing.T) {
	m := NewResponseMap[uint64, *mockResponse](1)

	go func() {
		time.Sleep(2 * time.Second)
		m.Set(1, &mockResponse{data: []byte("1")})
	}()

	res := m.Wait(1)
	assert.Equal(t, []byte("1"), res.data)
	assert.Nil(t, res.err)
}

func TestResponseMap_Wait_Timeout(t *testing.T) {
	m := NewResponseMap[uint64, *mockResponse](1)

	go func() {
		time.Sleep(3 * time.Second)
		m.Set(1, &mockResponse{data: []byte("1")})
	}()

	res := m.Wait(1)
	assert.Nil(t, res)
}
