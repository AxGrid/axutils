package comparer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type mock1 struct {
	Id uint64
}

func (m *mock1) GetId() uint64 {
	return m.Id
}

type mock2 struct {
	Id uint64
}

func (m *mock2) GetId() uint64 {
	return m.Id
}

func TestCompareSlice(t *testing.T) {
	mock1sl := []*mock1{
		{
			Id: 1,
		},
		{
			Id: 2,
		},
		{
			Id: 3,
		},
		{
			Id: 4,
		},
		{
			Id: 5,
		},
	}

	mock2sl := []*mock2{
		{
			Id: 1,
		},
		{
			Id: 3,
		},
		{
			Id: 5,
		},
		{
			Id: 7,
		},
	}

	del, create, modify := CompareSlice(mock1sl, mock2sl)
	assert.Equal(t, len(del), 2)
	assert.Equal(t, len(create), 1)
	assert.Equal(t, len(modify), 3)
}
