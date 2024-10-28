package collections

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestRequestMap_GetOrCreate(t *testing.T) {
	rm := NewRequestMap[int, string](time.Millisecond * 200)
	workCount := int32(0)
	longFunc := func(k int) string {
		t.Log("longFunc")
		time.Sleep(time.Millisecond * 100)
		atomic.AddInt32(&workCount, 1)
		return fmt.Sprintf("long:%d", k)
	}

	shortFunc := func(k int) string {
		t.Log("shortFunc")
		atomic.AddInt32(&workCount, 1)
		return fmt.Sprintf("short:%d", k)
	}

	go rm.GetOrCreate(1, longFunc)
	go rm.GetOrCreate(1, longFunc)
	go rm.GetOrCreate(2, shortFunc)
	go rm.GetOrCreate(1, longFunc)
	assert.Equal(t, rm.GetOrCreate(1, longFunc), "long:1")
	assert.Equal(t, rm.GetOrCreate(2, shortFunc), "short:2")
	assert.Equal(t, int32(2), workCount)
	assert.Equal(t, rm.GetOrCreate(2, shortFunc), "short:2")
	assert.Equal(t, int32(2), workCount)
	time.Sleep(time.Millisecond * 300)
	assert.Equal(t, rm.GetOrCreate(2, shortFunc), "short:2")
	assert.Equal(t, int32(3), workCount)
}
