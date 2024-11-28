package collections

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand/v2"
	"runtime"
	"sync"
	"testing"
	"time"
)

/*
   ________  ________   _______   ______    ______   _______    _______
  /        \/        \//       \//      \ //      \ /       \\//       \
 /        _/        _//        //       ///       //        ///        /
/-        //       //        _/        //        //         /        _/
\_______// \_____// \________/\________/\________/\___/____/\____/___/
zed (03.11.2024)
*/

func TestWaitMap_Has(t *testing.T) {
	type demo struct {
		name string
	}
	m := map[int]*demo{}
	m[1] = &demo{name: "demo"}
	m[2] = nil
	if _, ok := m[1]; !ok {
		t.Fatal("expected true")
	}
	if _, ok := m[2]; !ok {
		t.Fatal("expected true")
	}
	if _, ok := m[3]; ok {
		t.Fatal("expected false")
	}
	t.Log("ok", fmt.Sprintf("%v", m))
}

func TestWaitMap_Set(t *testing.T) {
	type demo struct {
		name string
	}
	wm := NewWaitMap[int, *demo]().Build()
	wm.Set(1, &demo{name: "demo"})
	assert.Equal(t, "demo", wm.Wait(1).name)
	wm.Set(1, &demo{name: "demo-bad"})
	assert.Equal(t, "demo", wm.Wait(1).name)
}

func TestWaitMap_WaitSet(t *testing.T) {
	type demo struct {
		name string
	}
	wm := NewWaitMap[int, *demo]().WithRequestTimeout(time.Millisecond * 50).WithResponseTtl(time.Millisecond * 300).Build()
	go func() {
		time.Sleep(time.Millisecond * 10)
		wm.Set(1, &demo{name: "demo"})
	}()

	go func() {
		time.Sleep(time.Millisecond * 100)
		wm.Set(2, &demo{name: "demo-w2"})
	}()
	w1 := wm.Wait(1)
	assert.Equal(t, "demo", w1.name)
	assert.Equal(t, 1, wm.Count())
	w2 := wm.Wait(2)
	assert.Equal(t, 2, wm.Count())

	assert.NotNil(t, w1)
	assert.Equal(t, "demo", w1.name)
	assert.Nil(t, w2)
	time.Sleep(time.Millisecond * 100)
	w1 = wm.Wait(1)
	w2 = wm.Wait(2)
	assert.Equal(t, 2, wm.Count())
	assert.NotNil(t, w1)
	assert.Equal(t, "demo", w1.name)
	assert.Nil(t, w2)
	time.Sleep(time.Millisecond * 200)
	assert.Equal(t, 0, wm.Count())
}

func TestWaitMap_MemoryLeak(t *testing.T) {
	type demo struct {
		name string
	}
	wm := NewWaitMap[int32, *demo]().WithRequestTimeout(time.Millisecond * 50).WithResponseTtl(time.Millisecond * 300).Build()
	wg := sync.WaitGroup{}

	gor := runtime.NumGoroutine()
	for i := 0; i < 1_000_000; i++ {
		wg.Add(1)
		r := rand.Int32N(3000)
		go func(trx int32) {
			time.Sleep(time.Millisecond * 10)
			wm.Set(trx, &demo{name: "demo"})
		}(r)
		go func() {
			wm.Wait(r)
			wg.Done()
		}()
	}
	wg.Wait()
	time.Sleep(time.Millisecond * 500)
	assert.Equal(t, 0, wm.Count())
	t.Log("delta goroutine", runtime.NumGoroutine()-gor)
	assert.Less(t, runtime.NumGoroutine(), gor+1)
}
