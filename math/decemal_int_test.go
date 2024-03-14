package math

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

func TestIntPointMove(t *testing.T) {
	a := 100
	assert.Equal(t, IntPointMove(a, 1), 1000)
	assert.Equal(t, IntPointMove(a, -1), 10)
}
