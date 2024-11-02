package collections

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHolderMap_Release(t *testing.T) {
	type HS struct {
		Object int
		Data   string
	}
	hm := NewHolderMap[int, *HS]().WithTimeout(time.Millisecond * 150).WithTTL(time.Millisecond * 30).Build()
	hO := &HS{
		Object: 1,
		Data:   "data",
	}
	err := hm.Wait(1, hO)
	assert.Error(t, err)
	err = hm.Wait(1, hO)
	assert.Error(t, err)

	go func() {
		h1 := &HS{
			Object: 2,
			Data:   "data-2",
		}
		err = hm.Wait(2, h1)
		assert.NoError(t, err)
		assert.Equal(t, h1.Data, "modificated")
		println("done: ", h1.Data)
	}()
	time.Sleep(time.Millisecond * 10)
	obj, err := hm.Get(2)
	assert.NoError(t, err)
	obj.Data = "modificated"
	assert.Nil(t, hm.Release(2))
	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, hm.Count(), 2)
	time.Sleep(time.Millisecond * 35)
	assert.Equal(t, hm.Count(), 0)

}
