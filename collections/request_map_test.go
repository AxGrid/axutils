package collections

import (
	"encoding/hex"
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

func TestRequestMap_Timeout(t *testing.T) {
	rm := NewRequestMap[int, string](time.Millisecond * 200)
	longFunc := func(k int) (string, error) {
		time.Sleep(time.Millisecond * 100)
		return fmt.Sprintf("long:%d", k), nil
	}
	tFunc := rm.Timeout(time.Millisecond*50, longFunc)
	v, err := tFunc(1)
	assert.ErrorIs(t, err, ErrTimeout)
	assert.Equal(t, "", v)
	time.Sleep(150 * time.Millisecond)
}

func TestRequestMap_GetOrCreateWeb(t *testing.T) {
	// Создаем где-то вверху на конструкторе
	rm := NewRequestMap[string, []byte](time.Minute * 20)
	for i := 0; i < 100000; i++ {
		go func(i int) {
			key := fmt.Sprintf("key-%d", i)
			data := rm.GetOrCreate(key, func(k string) []byte {
				return []byte(hex.EncodeToString([]byte(k)))
			})
			assert.Equal(t, data, []byte(hex.EncodeToString([]byte(key))))
		}(i)
	}

}
