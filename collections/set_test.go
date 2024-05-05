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
zed (04.05.2024)
*/

func TestSet(t *testing.T) {
	setA := NewSet[int]()
	setA.Add(1)
	setA.Add(2)
	setA.Add(3)
	setA.Add(1)
	assert.Equal(t, 3, setA.Size())
	assert.True(t, setA.Has(1))
	assert.True(t, setA.Has(2))
	assert.True(t, setA.Has(3))
	assert.False(t, setA.Has(4))
	setA.Remove(2)
	assert.Equal(t, 2, setA.Size())
	assert.False(t, setA.Has(2))

	setB := NewSet[int]()
	setB.Add(3)
	setB.Add(1)
	assert.Equal(t, setA, setB)
	assert.Equal(t, setA.Values(), setB.Values())
	setB.Add(4)
	assert.NotEqual(t, setA, setB)
	assert.NotEqual(t, setA.Values(), setB.Values())
	setA.Add(4)
	assert.Equal(t, setA, setB)
}
