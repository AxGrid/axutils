package collections

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
		t.Log("build", key)
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
