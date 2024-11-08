package collections

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand/v2"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRequestMap_GetOrCreate(t *testing.T) {
	rm := NewRequestMap[int, string](context.Background(), time.Millisecond*200)
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
	rm := NewRequestMap[int, string](context.Background(), time.Millisecond*200)
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
	rm := NewRequestMap[string, []byte](context.Background(), time.Minute*20)
	for i := 0; i < 100000; i++ {
		go func(i int) {
			key := fmt.Sprintf("Key-%d", i)
			data := rm.GetOrCreate(key, func(k string) []byte {
				return []byte(hex.EncodeToString([]byte(k)))
			})
			assert.Equal(t, data, []byte(hex.EncodeToString([]byte(key))))
		}(i)
	}

}

func TestWaitMap_Billing(t *testing.T) {
	type demo struct {
		value  int
		err    error
		errMsg string
	}
	billingFunc := func(trx string) *demo {
		time.Sleep(time.Millisecond * 300)
		if rand.Int32N(100) < 20 {
			return &demo{value: 0, err: fmt.Errorf("error"), errMsg: "error:" + trx}
		}
		r := rand.Int32N(100000)
		return &demo{value: int(r)}
	}

	wa := NewRequestMap[string, *demo](context.Background(), time.Minute*30)
	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			trx := fmt.Sprintf("trx-%d", rand.Int32N(20))
			res := wa.GetOrCreate(trx, func(trx string) *demo {
				res := billingFunc(trx)
				return res
			})
			t.Log(trx, "Result", res.value, res.err, res.errMsg)
		}(i)
	}
	wg.Wait()
}

func TestWaitMap_Billing2(t *testing.T) {
	type demo struct {
		value  int
		err    error
		errMsg string
	}
	billingFunc := func(trx string) *demo {
		t.Log("REQUEST", trx)
		time.Sleep(time.Millisecond * 300)

		if rand.Int32N(100) < 20 {
			return &demo{value: 0, err: fmt.Errorf("error"), errMsg: "error:" + trx}
		}
		r := rand.Int32N(100000)
		t.Log("RESPONSE", trx)
		return &demo{value: int(r)}
	}

	// Достаем последние запросы из БД (по created_at) (time.Minute * 30)
	prevsResult := make([]*RequestMapInitializer[string, *demo], 0, 10)
	for i := 1; i < 10; i++ {
		trx := fmt.Sprintf("trx-%d", i)
		prevsResult = append(prevsResult, &RequestMapInitializer[string, *demo]{Key: trx, Result: &demo{value: 12000 + i}})
	}

	// Создаем очередь запросов
	wa := NewRequestMap[string, *demo](context.Background(), time.Minute*30, prevsResult...)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			trx := fmt.Sprintf("trx-%d", 5)
			t.Log(time.Now().Format(time.StampMicro), trx, "request")
			res := wa.GetOrCreate(trx, func(trx string) *demo {
				res := billingFunc(trx)
				return res
			})
			t.Log(time.Now().Format(time.StampMicro), trx, "Result", res.value, res.err, res.errMsg)
		}(i)
	}
	wg.Wait()
	t.Log("-------------------")
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			trx := fmt.Sprintf("trx-%d", 5)
			t.Log(time.Now().Format(time.StampMicro), trx, "request")
			res := wa.GetOrCreate(trx, func(trx string) *demo {
				res := billingFunc(trx)
				return res
			})
			t.Log(time.Now().Format(time.StampMicro), trx, "Result", res.value, res.err, res.errMsg)
		}(i)
	}

	wg.Wait()

}

func TestWaitMap_Billing4(t *testing.T) {
	type demo struct {
		value  int
		err    error
		errMsg string
	}
	wa := NewRequestMap[string, *demo](context.Background(), time.Millisecond*50)
	r := wa.GetOrCreate("trx-1", func(trx string) *demo {
		t.Log("create 1")
		time.Sleep(time.Millisecond * 100)
		return &demo{value: 100}
	})
	t.Log(r.value)
	assert.Equal(t, 100, r.value)
	time.Sleep(time.Millisecond * 20)
	r = wa.GetOrCreate("trx-1", func(trx string) *demo {
		t.Log("create 1")
		time.Sleep(time.Millisecond * 100)
		return &demo{value: 100}
	})
	t.Log(r.value)
	assert.Equal(t, 100, r.value)
	time.Sleep(time.Millisecond * 200)

	r = wa.GetOrCreate("trx-1", func(trx string) *demo {
		time.Sleep(time.Millisecond * 100)
		t.Log("create 2")
		return &demo{value: 150}
	})
	t.Log(r.value)
	assert.Equal(t, 150, r.value)
}

func TestWaitMap_Billing3(t *testing.T) {
	type demo struct {
		value  int
		err    error
		errMsg string
	}
	billingFunc := func(trx string) *demo {
		time.Sleep(time.Millisecond * 400)
		if rand.Int32N(100) < 20 {
			return &demo{value: 0, err: fmt.Errorf("error"), errMsg: "error:" + trx}
		}
		r := rand.Int32N(100000)
		return &demo{value: int(r)}
	}
	startTime := time.Now()
	startGor := runtime.NumGoroutine()
	resMap := make(map[string]*demo)
	resMapMu := sync.RWMutex{}
	// Создаем очередь запросов
	wa := NewRequestMap[string, *demo](context.Background(), time.Millisecond*300)
	wg := sync.WaitGroup{}
	t.Log(time.Now().Format(time.StampMicro), "start")
	for i := 0; i < 1_000_000; i++ {
		wg.Add(1)
		trx := fmt.Sprintf("trx-%d", rand.Int32N(1000))
		go func(i int) {
			if i < 10 {
				t.Log(time.Now().Format(time.StampMicro), "start", trx)
			}
			defer wg.Done()
			time.Sleep(time.Duration(rand.Int32N(100)) * time.Millisecond)

			res := wa.GetOrCreate(trx, func(trx string) *demo {
				res := billingFunc(trx)
				return res
			})
			resMapMu.Lock()
			resMap[trx] = res
			resMapMu.Unlock()
		}(i)
		if i < 10 {
			t.Log(time.Now().Format(time.StampMicro), "start-i")
		}
	}

	go func() {
		for {
			time.Sleep(time.Second)
			resMapMu.RLock()
			t.Log("count-map", wa.Count(), len(resMap))
			resMapMu.RUnlock()
		}
	}()

	wg.Wait()
	time.Sleep(time.Millisecond * 10)
	t.Log("count", wa.Count())
	t.Log("delta goroutine", runtime.NumGoroutine()-startGor)
	t.Log("delta time", time.Now().Sub(startTime).Milliseconds())
	time.Sleep(time.Millisecond * 500)
	t.Log("count", wa.Count())
	t.Log("delta goroutine", runtime.NumGoroutine()-startGor)
}
