package flow

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (05.05.2024)
*/

type TestState struct {
	state string
}

func (t *TestState) GetState() string {
	return t.state
}

func (t *TestState) SetState(s string) {
	t.state = s
}

func TestFlowProcess(t *testing.T) {
	flow := &Flow[int, string, *TestState]{}
	state := TestState{state: "init"}
	err := flow.Process(&state, 1)
	assert.Nil(t, err)
}
