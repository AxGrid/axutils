package fms

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (09.04.2024)
*/

func TestFlowBuilder(t *testing.T) {
	f := NewFlow[string, *int]().
		Route(func(rb *FlowProcessorRouterBuilder[string, *int]) {
			// Есил пришло событие "test" то увеличиваем state на 1
			rb.OnEvent(func(event string) bool {
				if event == "test" {
					return true
				}
				return false
			}).Do(func(_ string, state *int) error {
				*state = *state + 1
				return nil
			}).Build()
		}).
		// Если пришло событие "test" то увеличиваем state на 5
		RouteOn(func(event string) bool {
			if event == "test" {
				return true
			}
			return false
		}, func(event string, state *int) error {
			*state = *state + 5
			return nil
		}).
		Build()
	var state int

	err := f("test", &state)
	assert.Nil(t, err)
	assert.Equal(t, 6, state)

}
