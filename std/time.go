package std

import "time"

func Microtime() float64 {
	return float64(time.Now().UnixNano()) / 1000000000
}
