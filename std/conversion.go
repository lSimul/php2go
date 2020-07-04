package std

// BoolToInt converts b to int
// the same way like PHP do.
func BoolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

// BoolToFloat converts b to float
// the same way like PHP do.
func BoolToFloat64(b bool) float64 {
	return float64(BoolToInt(b))
}
