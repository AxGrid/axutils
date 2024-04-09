package flow2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewFlow_Success(t *testing.T) {
	f := NewFlow[string, *int]().
		Route(func(r *FlowProcessorRouterBuilder[string, *int]) {
			r.OnEvent(func(e string) bool {
				if e == "test" {
					return true
				}
				return false
			}).Do(func(event string, state *int) error {
				*state = *state + 1
				return nil
			}).Build()
		}).
		RouteOn(func(event string) bool {
			if event == "test" {
				return true
			}
			return false
		}, func(event string, state *int) error {
			*state = *state + 5
			return nil
		}).Build()

	var state int
	err := f("test", &state)
	assert.Nil(t, err)
	assert.Equal(t, 6, state)
}

func TestNewFlow_Errors(t *testing.T) {
	f := NewFlow[string, *int]().
		RouteOn(func(event string) bool {
			if event == "test" {
				return true
			}
			return false
		}, func(event string, state *int) error {
			*state = *state + 5
			return nil
		}).Build()

	var state int
	err := f("abc", &state)
	assert.NotNil(t, err)
	assert.Equal(t, 0, state)
}
