package math

/*
 __    _           ___
|  |  |_|_____ ___|_  |
|  |__| |     | .'|  _|
|_____|_|_|_|_|__,|___|
zed (14.03.2024)
*/

func IntPointMove(value int, shift int) int {
	if shift == 0 {
		return value
	}
	if shift > 0 {
		for i := 0; i < shift; i++ {
			value *= 10
		}
	} else {
		for i := 0; i < -shift; i++ {
			value /= 10
		}
	}
	return value
}
