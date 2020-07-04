package std

import "time"

// Microtime returns current Unix timestamp
// with microseconds.
// It represents PHP function with the same name,
// with one big difference, this value is always
// float.
//
// See php.net/manual/en/function.microtime.php
// for more details.
func Microtime() float64 {
	return float64(time.Now().UnixNano()) / 1000000000
}
